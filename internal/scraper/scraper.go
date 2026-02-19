package scraper

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/thinkscotty/kibble/internal/ai"
	"github.com/thinkscotty/kibble/internal/models"
	"github.com/thinkscotty/kibble/internal/reddit"
)

// Scraper handles web scraping operations.
type Scraper struct {
	userAgent      string
	requestTimeout time.Duration
	parallelLimit  int
	redditClient   *reddit.Client
}

// ScrapeResult represents the result of scraping a single source.
type ScrapeResult struct {
	Source  models.NewsSource
	Content *ai.ScrapedContent
	Error   error
}

// New creates a new Scraper.
func New() *Scraper {
	return &Scraper{
		userAgent:      "Kibble/1.0 (AI Facts & News Dashboard; +https://github.com/thinkscotty/kibble)",
		requestTimeout: 30 * time.Second,
		parallelLimit:  5,
		redditClient:   reddit.New(),
	}
}

// ScrapeSource scrapes content from a single source.
func (s *Scraper) ScrapeSource(ctx context.Context, source models.NewsSource) (*ai.ScrapedContent, error) {
	if reddit.IsRedditURL(source.URL) {
		return s.scrapeRedditSource(ctx, source)
	}

	// Try RSS/Atom feed parsing for URLs that look like feeds.
	// This uses encoding/xml which properly handles XML content,
	// unlike Colly's HTML parser which mangles RSS/Atom XML.
	if isRSSURL(source.URL) {
		content, err := s.scrapeRSSFeed(ctx, source)
		if err == nil {
			return content, nil
		}
		slog.Debug("RSS feed parsing failed, falling back to HTML scraping",
			"url", source.URL, "error", err)
	}

	c := colly.NewCollector(
		colly.UserAgent(s.userAgent),
		colly.MaxDepth(1),
	)
	c.SetRequestTimeout(s.requestTimeout)

	var content strings.Builder
	var title string
	var mu sync.Mutex

	c.OnHTML("title", func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()
		if title == "" {
			title = strings.TrimSpace(e.Text)
		}
	})

	contentSelectors := []string{
		"article", "main", ".content", ".post",
		".article", ".entry-content", "#content", "#main",
	}
	for _, selector := range contentSelectors {
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			mu.Lock()
			defer mu.Unlock()
			text := cleanText(e.Text)
			if len(text) > 100 {
				content.WriteString(text)
				content.WriteString("\n\n")
			}
		})
	}

	c.OnHTML("h1, h2, h3", func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()
		text := cleanText(e.Text)
		if len(text) > 10 && len(text) < 200 {
			content.WriteString("HEADLINE: ")
			content.WriteString(text)
			content.WriteString("\n")
		}
	})

	c.OnHTML("p", func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()
		text := cleanText(e.Text)
		if len(text) > 50 && len(text) < 2000 {
			content.WriteString(text)
			content.WriteString("\n")
		}
	})

	c.OnHTML("item, entry", func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()
		itemTitle := e.ChildText("title")
		itemDesc := e.ChildText("description, summary, content")
		itemLink := e.ChildAttr("link", "href")
		if itemLink == "" {
			itemLink = e.ChildText("link")
		}
		if itemTitle != "" {
			content.WriteString("ARTICLE: ")
			content.WriteString(itemTitle)
			content.WriteString("\n")
			if itemLink != "" {
				content.WriteString("LINK: ")
				content.WriteString(itemLink)
				content.WriteString("\n")
			}
			if itemDesc != "" {
				content.WriteString(cleanText(itemDesc))
				content.WriteString("\n\n")
			}
		}
	})

	var scrapeErr error
	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("scrape error for %s: %w (status: %d)", source.URL, err, r.StatusCode)
	})

	if err := c.Visit(source.URL); err != nil {
		return nil, fmt.Errorf("failed to visit %s: %w", source.URL, err)
	}
	c.Wait()

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	contentStr := content.String()
	if len(contentStr) < 100 {
		return nil, fmt.Errorf("insufficient content scraped from %s", source.URL)
	}

	const maxLength = 50000
	if len(contentStr) > maxLength {
		contentStr = contentStr[:maxLength] + "..."
	}

	sourceName := source.Name
	if sourceName == "" {
		sourceName = title
	}
	if sourceName == "" {
		if parsed, err := url.Parse(source.URL); err == nil {
			sourceName = parsed.Host
		}
	}

	return &ai.ScrapedContent{
		URL:        source.URL,
		SourceName: sourceName,
		Content:    contentStr,
	}, nil
}

// ScrapeSources scrapes multiple sources concurrently.
func (s *Scraper) ScrapeSources(ctx context.Context, sources []models.NewsSource) []ScrapeResult {
	var results []ScrapeResult
	var mu sync.Mutex

	sem := make(chan struct{}, s.parallelLimit)
	var wg sync.WaitGroup

	for _, source := range sources {
		select {
		case <-ctx.Done():
			return results
		default:
		}

		wg.Add(1)
		go func(src models.NewsSource) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					results = append(results, ScrapeResult{
						Source: src,
						Error:  fmt.Errorf("panic while scraping: %v", r),
					})
					mu.Unlock()
				}
			}()

			sem <- struct{}{}
			defer func() { <-sem }()

			content, err := s.ScrapeSource(ctx, src)

			mu.Lock()
			results = append(results, ScrapeResult{
				Source:  src,
				Content: content,
				Error:   err,
			})
			mu.Unlock()
		}(source)
	}

	wg.Wait()
	return results
}

// ValidateURL checks if a URL is valid and uses http/https.
func ValidateURL(urlStr string) error {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}

func (s *Scraper) scrapeRedditSource(ctx context.Context, source models.NewsSource) (*ai.ScrapedContent, error) {
	posts, err := s.redditClient.FetchPosts(ctx, source.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Reddit posts: %w", err)
	}
	if len(posts) == 0 {
		return nil, fmt.Errorf("no valid posts found in subreddit (text posts with >100 words)")
	}

	var content strings.Builder
	for _, post := range posts {
		fmt.Fprintf(&content, "REDDIT POST: %s\n", post.Title)
		fmt.Fprintf(&content, "LINK: https://reddit.com%s\n", post.Permalink)
		fmt.Fprintf(&content, "SCORE: %d | AUTHOR: u/%s\n", post.Score, post.Author)
		content.WriteString(post.Body)
		content.WriteString("\n\n---\n\n")
	}

	contentStr := content.String()
	const maxLength = 50000
	if len(contentStr) > maxLength {
		contentStr = contentStr[:maxLength] + "..."
	}

	sourceName := source.Name
	if sourceName == "" {
		sourceName = extractSubredditName(source.URL)
	}

	return &ai.ScrapedContent{
		URL:        source.URL,
		SourceName: sourceName,
		Content:    contentStr,
	}, nil
}

func extractSubredditName(u string) string {
	if idx := strings.Index(u, "/r/"); idx != -1 {
		rest := u[idx+3:]
		if slashIdx := strings.Index(rest, "/"); slashIdx != -1 {
			return "r/" + rest[:slashIdx]
		}
		return "r/" + rest
	}
	return "reddit"
}

func cleanText(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// --- RSS/Atom feed parsing ---

// isRSSURL checks if a URL looks like an RSS or Atom feed based on path patterns.
func isRSSURL(u string) bool {
	lower := strings.ToLower(u)
	// Strip query string and fragment for path-based checks
	if idx := strings.IndexAny(lower, "?#"); idx >= 0 {
		lower = lower[:idx]
	}
	lower = strings.TrimRight(lower, "/")

	return strings.HasSuffix(lower, "/feed") ||
		strings.HasSuffix(lower, "/rss") ||
		strings.HasSuffix(lower, "/atom") ||
		strings.HasSuffix(lower, ".xml") ||
		strings.HasSuffix(lower, ".rss") ||
		strings.HasSuffix(lower, ".atom") ||
		strings.Contains(lower, "/feeds/") ||
		strings.Contains(lower, "/rss/")
}

// RSS 2.0 XML types
type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title          string `xml:"title"`
	Link           string `xml:"link"`
	Description    string `xml:"description"`
	PubDate        string `xml:"pubDate"`
	ContentEncoded string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
}

// Atom XML types
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title   string     `xml:"title"`
	Links   []atomLink `xml:"link"`
	Summary string     `xml:"summary"`
	Content string     `xml:"content"`
	Updated string     `xml:"updated"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// scrapeRSSFeed fetches and parses an RSS/Atom feed, returning structured content.
func (s *Scraper) scrapeRSSFeed(ctx context.Context, source models.NewsSource) (*ai.ScrapedContent, error) {
	client := &http.Client{Timeout: s.requestTimeout}

	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed %s: %w", source.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("feed returned status %d for %s", resp.StatusCode, source.URL)
	}

	// If the server explicitly returns HTML, this isn't a feed
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("URL returned HTML content-type, not a feed")
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("read feed body: %w", err)
	}

	// Try RSS 2.0
	var rss rssFeed
	if xml.Unmarshal(body, &rss) == nil && len(rss.Channel.Items) > 0 {
		slog.Info("Parsed RSS feed", "url", source.URL, "items", len(rss.Channel.Items),
			"title", rss.Channel.Title)
		return formatRSSItems(source, rss.Channel.Title, rss.Channel.Items), nil
	}

	// Try Atom
	var atom atomFeed
	if xml.Unmarshal(body, &atom) == nil && len(atom.Entries) > 0 {
		slog.Info("Parsed Atom feed", "url", source.URL, "entries", len(atom.Entries),
			"title", atom.Title)
		return formatAtomEntries(source, atom.Title, atom.Entries), nil
	}

	return nil, fmt.Errorf("URL %s is not a recognized RSS/Atom feed", source.URL)
}

func formatRSSItems(source models.NewsSource, feedTitle string, items []rssItem) *ai.ScrapedContent {
	var content strings.Builder
	for _, item := range items {
		if item.Title == "" {
			continue
		}
		content.WriteString("ARTICLE: ")
		content.WriteString(item.Title)
		content.WriteString("\n")
		if item.Link != "" {
			content.WriteString("LINK: ")
			content.WriteString(item.Link)
			content.WriteString("\n")
		}
		if item.PubDate != "" {
			content.WriteString("DATE: ")
			content.WriteString(item.PubDate)
			content.WriteString("\n")
		}
		// Prefer content:encoded (full article) over description (summary)
		desc := item.ContentEncoded
		if desc == "" {
			desc = item.Description
		}
		if desc != "" {
			content.WriteString(cleanText(stripHTMLTags(desc)))
			content.WriteString("\n\n")
		}
	}

	return buildScrapedContent(source, feedTitle, content.String())
}

func formatAtomEntries(source models.NewsSource, feedTitle string, entries []atomEntry) *ai.ScrapedContent {
	var content strings.Builder
	for _, entry := range entries {
		if entry.Title == "" {
			continue
		}
		content.WriteString("ARTICLE: ")
		content.WriteString(entry.Title)
		content.WriteString("\n")
		if link := atomEntryLink(entry); link != "" {
			content.WriteString("LINK: ")
			content.WriteString(link)
			content.WriteString("\n")
		}
		if entry.Updated != "" {
			content.WriteString("DATE: ")
			content.WriteString(entry.Updated)
			content.WriteString("\n")
		}
		// Prefer content over summary
		desc := entry.Content
		if desc == "" {
			desc = entry.Summary
		}
		if desc != "" {
			content.WriteString(cleanText(stripHTMLTags(desc)))
			content.WriteString("\n\n")
		}
	}

	return buildScrapedContent(source, feedTitle, content.String())
}

// atomEntryLink extracts the best link from an Atom entry.
func atomEntryLink(entry atomEntry) string {
	// Prefer rel="alternate", fall back to first link
	for _, l := range entry.Links {
		if l.Rel == "alternate" || l.Rel == "" {
			return l.Href
		}
	}
	if len(entry.Links) > 0 {
		return entry.Links[0].Href
	}
	return ""
}

func buildScrapedContent(source models.NewsSource, feedTitle, contentStr string) *ai.ScrapedContent {
	const maxLength = 50000
	if len(contentStr) > maxLength {
		contentStr = contentStr[:maxLength] + "..."
	}

	sourceName := source.Name
	if sourceName == "" {
		sourceName = feedTitle
	}
	if sourceName == "" {
		if parsed, _ := url.Parse(source.URL); parsed != nil {
			sourceName = parsed.Host
		}
	}

	return &ai.ScrapedContent{
		URL:        source.URL,
		SourceName: sourceName,
		Content:    contentStr,
	}
}

// stripHTMLTags removes HTML tags from a string. RSS feed content often contains
// HTML markup in description and content:encoded elements.
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			result.WriteRune(' ')
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}
