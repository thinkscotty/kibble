package scraper

import (
	"context"
	"fmt"
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
