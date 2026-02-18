package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/thinkscotty/kibble/internal/models"
	"github.com/thinkscotty/kibble/internal/reddit"
)

// ValidationResult holds the result of a source validation attempt.
type ValidationResult struct {
	URL     string
	Name    string
	OK      bool
	Reason  string // why it failed, if !OK
	FeedURL string // RSS feed URL, if discovered during validation
}

// ValidateSource performs a lightweight test-scrape of a source URL to confirm
// it returns usable content. If the source is a web page (not Reddit), it also
// attempts RSS feed auto-discovery and prefers the feed URL if found.
func (s *Scraper) ValidateSource(ctx context.Context, sourceURL, name string) ValidationResult {
	result := ValidationResult{URL: sourceURL, Name: name}

	testURL := sourceURL

	// Try RSS discovery for non-Reddit URLs
	if !reddit.IsRedditURL(sourceURL) {
		if feedURL := DiscoverRSSFeed(ctx, sourceURL); feedURL != "" {
			result.FeedURL = feedURL
			testURL = feedURL // validate the feed URL instead
		}
	}

	// Create a temporary NewsSource for ScrapeSource
	testSource := models.NewsSource{
		URL:  testURL,
		Name: name,
	}

	// Use a shorter timeout for validation
	valCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	content, err := s.ScrapeSource(valCtx, testSource)
	if err != nil {
		result.OK = false
		result.Reason = err.Error()
		return result
	}

	if len(content.Content) < 200 {
		result.OK = false
		result.Reason = fmt.Sprintf("insufficient content: %d chars", len(content.Content))
		return result
	}

	result.OK = true
	return result
}
