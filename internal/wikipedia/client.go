package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client queries the Wikipedia API for search results and article summaries.
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// SearchResult represents a single Wikipedia search hit.
type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	PageID  int    `json:"pageid"`
}

// New creates a Wikipedia client with a 15-second timeout.
func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		userAgent:  "Kibble/1.0 (AI Facts Dashboard; +https://github.com/thinkscotty/kibble)",
	}
}

// Search finds Wikipedia articles matching a query.
func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	params := url.Values{
		"action":  {"query"},
		"list":    {"search"},
		"srsearch": {query},
		"format":  {"json"},
		"utf8":    {"1"},
		"srlimit": {fmt.Sprintf("%d", limit)},
	}

	reqURL := "https://en.wikipedia.org/w/api.php?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wikipedia search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikipedia search returned %d", resp.StatusCode)
	}

	var result struct {
		Query struct {
			Search []SearchResult `json:"search"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	return result.Query.Search, nil
}

// GetSummary fetches a concise article summary using the REST API.
func (c *Client) GetSummary(ctx context.Context, title string) (string, error) {
	encoded := url.PathEscape(strings.ReplaceAll(title, " ", "_"))
	reqURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", encoded)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("wikipedia summary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("wikipedia summary returned %d for %q", resp.StatusCode, title)
	}

	var result struct {
		Title   string `json:"title"`
		Extract string `json:"extract"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode summary response: %w", err)
	}

	if result.Extract == "" {
		return "", fmt.Errorf("no summary available for %q", title)
	}

	return fmt.Sprintf("## %s\n%s", result.Title, result.Extract), nil
}
