package scraper

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// DiscoverRSSFeed checks a web page for RSS/Atom feed <link> tags.
// Returns the feed URL if found, or empty string if none discovered.
func DiscoverRSSFeed(ctx context.Context, pageURL string) string {
	c := colly.NewCollector(
		colly.UserAgent("Kibble/1.0 (RSS Discovery; +https://github.com/thinkscotty/kibble)"),
		colly.MaxDepth(0),
	)
	c.SetRequestTimeout(10 * time.Second)

	var feedURL string
	var mu sync.Mutex

	c.OnHTML(`link[rel="alternate"]`, func(e *colly.HTMLElement) {
		mu.Lock()
		defer mu.Unlock()
		if feedURL != "" {
			return // already found one
		}
		typ := strings.ToLower(e.Attr("type"))
		if typ == "application/rss+xml" || typ == "application/atom+xml" {
			href := e.Attr("href")
			if href != "" {
				feedURL = resolveURL(pageURL, href)
			}
		}
	})

	c.Visit(pageURL)
	c.Wait()

	return feedURL
}

// resolveURL resolves a potentially relative href against a base URL.
func resolveURL(base, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return href
	}

	ref, err := url.Parse(href)
	if err != nil {
		return href
	}

	return baseURL.ResolveReference(ref).String()
}
