package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Client handles fetching posts from Reddit's JSON API.
type Client struct {
	httpClient  *http.Client
	userAgent   string
	minWords    int
	mu          sync.Mutex
	lastRequest time.Time
	minInterval time.Duration
}

// Post represents a filtered Reddit post.
type Post struct {
	Title      string
	Body       string
	Permalink  string
	Subreddit  string
	Author     string
	Score      int
	CreatedUTC time.Time
}

// New creates a new Reddit client with rate limiting.
func New() *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		userAgent:   "Kibble/1.0 (AI Facts & News Dashboard; +https://github.com/thinkscotty/kibble)",
		minWords:    100,
		minInterval: 1100 * time.Millisecond,
	}
}

// FetchPosts fetches and filters text posts from a subreddit.
func (c *Client) FetchPosts(ctx context.Context, subredditURL string) ([]Post, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	subreddit, err := extractSubreddit(subredditURL)
	if err != nil {
		return nil, err
	}

	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://www.reddit.com/r/%s.json?limit=25", subreddit)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch subreddit %s: %w", subreddit, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("subreddit r/%s not found", subreddit)
		case http.StatusForbidden:
			return nil, fmt.Errorf("subreddit r/%s is private or quarantined", subreddit)
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("Reddit rate limit exceeded")
		default:
			return nil, fmt.Errorf("Reddit API returned status %d", resp.StatusCode)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var listing redditListing
	if err := json.Unmarshal(body, &listing); err != nil {
		return nil, fmt.Errorf("parse Reddit JSON: %w", err)
	}

	var posts []Post
	for _, child := range listing.Data.Children {
		post := child.Data
		if !post.IsSelf {
			continue
		}
		if len(strings.Fields(post.Selftext)) < c.minWords {
			continue
		}
		posts = append(posts, Post{
			Title:      post.Title,
			Body:       post.Selftext,
			Permalink:  post.Permalink,
			Subreddit:  post.Subreddit,
			Author:     post.Author,
			Score:      post.Score,
			CreatedUTC: time.Unix(int64(post.CreatedUTC), 0),
		})
	}

	return posts, nil
}

func (c *Client) waitForRateLimit(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.lastRequest)
	if elapsed < c.minInterval {
		wait := c.minInterval - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	c.lastRequest = time.Now()
	return nil
}

// IsRedditURL checks if a URL is a Reddit subreddit URL.
func IsRedditURL(url string) bool {
	return strings.Contains(url, "reddit.com/r/") || strings.HasPrefix(url, "r/")
}

func extractSubreddit(url string) (string, error) {
	patterns := []string{
		`reddit\.com/r/([a-zA-Z0-9_]+)`,
		`^r/([a-zA-Z0-9_]+)`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("could not extract subreddit from URL: %s", url)
}

// Reddit JSON API types

type redditListing struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPost struct {
	Title      string  `json:"title"`
	Selftext   string  `json:"selftext"`
	IsSelf     bool    `json:"is_self"`
	Permalink  string  `json:"permalink"`
	Subreddit  string  `json:"subreddit"`
	Author     string  `json:"author"`
	Score      int     `json:"score"`
	CreatedUTC float64 `json:"created_utc"`
}
