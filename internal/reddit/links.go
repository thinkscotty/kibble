package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// LinkPost represents a Reddit link post pointing to an external URL.
type LinkPost struct {
	Title     string
	URL       string
	Domain    string
	Subreddit string
	Permalink string
	Score     int
}

// DomainRank represents a domain found in subreddit link posts, ranked by frequency and score.
type DomainRank struct {
	Domain     string
	Count      int
	TotalScore int
	SampleURLs []string // up to 3 example article URLs from this domain
}

// FetchTopLinks fetches link posts (not self-posts) from a subreddit's top posts of the week.
// These are external links that the community has shared and upvoted.
func (c *Client) FetchTopLinks(ctx context.Context, subredditURL string, limit int) ([]LinkPost, error) {
	subreddit, err := extractSubreddit(subredditURL)
	if err != nil {
		return nil, err
	}

	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://www.reddit.com/r/%s/top.json?t=week&limit=%d", subreddit, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch top links from r/%s: %w", subreddit, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Reddit API returned status %d for r/%s/top", resp.StatusCode, subreddit)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var listing redditListing
	if err := json.Unmarshal(body, &listing); err != nil {
		return nil, fmt.Errorf("parse Reddit JSON: %w", err)
	}

	var posts []LinkPost
	for _, child := range listing.Data.Children {
		post := child.Data
		if post.IsSelf {
			continue // skip self-posts, we want external links
		}
		if post.Score < 10 {
			continue // skip low-score posts
		}
		if isMediaDomain(post.Domain) {
			continue // skip image/video hosts
		}
		posts = append(posts, LinkPost{
			Title:     post.Title,
			URL:       post.URL,
			Domain:    post.Domain,
			Subreddit: post.Subreddit,
			Permalink: post.Permalink,
			Score:     post.Score,
		})
	}

	return posts, nil
}

// RankDomains aggregates link posts by domain and ranks them by frequency then total score.
func RankDomains(posts []LinkPost) []DomainRank {
	domainMap := make(map[string]*DomainRank)

	for _, post := range posts {
		domain := normalizeDomain(post.Domain)
		rank, exists := domainMap[domain]
		if !exists {
			rank = &DomainRank{Domain: domain}
			domainMap[domain] = rank
		}
		rank.Count++
		rank.TotalScore += post.Score
		if len(rank.SampleURLs) < 3 {
			rank.SampleURLs = append(rank.SampleURLs, post.URL)
		}
	}

	var ranks []DomainRank
	for _, r := range domainMap {
		ranks = append(ranks, *r)
	}

	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].Count != ranks[j].Count {
			return ranks[i].Count > ranks[j].Count
		}
		return ranks[i].TotalScore > ranks[j].TotalScore
	})

	return ranks
}

// isMediaDomain checks if a domain is an image/video host that shouldn't be used as a news source.
func isMediaDomain(domain string) bool {
	mediaDomains := []string{
		"i.redd.it", "v.redd.it", "imgur.com", "i.imgur.com",
		"youtube.com", "youtu.be", "gfycat.com", "streamable.com",
		"twitter.com", "x.com", "reddit.com",
	}
	domain = strings.ToLower(domain)
	for _, m := range mediaDomains {
		if domain == m || strings.HasSuffix(domain, "."+m) {
			return true
		}
	}
	return false
}

func normalizeDomain(domain string) string {
	domain = strings.ToLower(domain)
	domain = strings.TrimPrefix(domain, "www.")
	return domain
}
