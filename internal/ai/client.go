package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/thinkscotty/kibble/internal/feeds"
	"github.com/thinkscotty/kibble/internal/wikipedia"
)

// Client is the main AI entry point. It routes requests to the correct provider
// and handles prompt building and response parsing.
type Client struct {
	gemini   *GeminiProvider
	ollama   *OllamaProvider
	settings SettingsGetter
	wiki     *wikipedia.Client
}

// NewClient creates an AI client with both providers and optional Wikipedia client.
func NewClient(sg SettingsGetter, wiki *wikipedia.Client) *Client {
	return &Client{
		gemini:   NewGeminiProvider(sg),
		ollama:   NewOllamaProvider(sg),
		settings: sg,
		wiki:     wiki,
	}
}

// resolveProvider returns the correct provider based on per-topic override or global setting.
// topicProvider: "" means use global default, "gemini" or "ollama" means use that provider.
func (c *Client) resolveProvider(topicProvider string) Provider {
	provider := topicProvider
	if provider == "" {
		provider, _ = c.settings.GetSetting("ai_provider")
	}

	switch provider {
	case "ollama":
		return c.ollama
	default:
		return c.gemini
	}
}

// GenerateFacts generates facts for a topic.
// If the topic is marked as niche and a Wikipedia client is available,
// it automatically performs research and uses a RAG-augmented prompt.
// Returns: facts, tokensUsed, providerName, modelName, error.
func (c *Client) GenerateFacts(ctx context.Context, opts FactsOpts) ([]string, int, string, string, error) {
	provider := c.resolveProvider(opts.AIProvider)

	var prompt string
	if opts.IsNiche && c.wiki != nil {
		researchCtx, err := c.ResearchTopic(ctx, provider, opts.Topic, opts.Description)
		if err != nil {
			slog.Warn("Wikipedia research failed, falling back to standard prompt", "topic", opts.Topic, "error", err)
		}
		if researchCtx != "" {
			prompt = BuildFactsPromptWithContext(
				opts.Topic, opts.Description,
				opts.CustomInstructions, opts.ToneInstructions,
				opts.Count, opts.MinWords, opts.MaxWords,
				researchCtx,
			)
		}
	}
	if prompt == "" {
		prompt = BuildFactsPrompt(
			opts.Topic, opts.Description,
			opts.CustomInstructions, opts.ToneInstructions,
			opts.Count, opts.MinWords, opts.MaxWords,
		)
	}

	resp, err := provider.Chat(ctx, ChatRequest{
		Messages:    []Message{{Role: "user", Content: prompt}},
		Temperature: 0.9,
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, 0, provider.Name(), "", err
	}

	facts := ParseFactsFromText(resp.Content)
	return facts, resp.TokensUsed, resp.Provider, resp.Model, nil
}

// DiscoverSources uses AI to find news sources for a topic.
// If the topic is marked as niche and a Wikipedia client is available,
// it automatically performs research and uses a RAG-augmented prompt.
func (c *Client) DiscoverSources(ctx context.Context, opts DiscoverOpts) ([]DiscoveredSource, int, string, string, error) {
	provider := c.resolveProvider(opts.AIProvider)

	suggested := feeds.FindRelevant(opts.TopicName, opts.Description)

	var prompt string
	if opts.IsNiche && c.wiki != nil {
		researchCtx, err := c.ResearchTopic(ctx, provider, opts.TopicName, opts.Description)
		if err != nil {
			slog.Warn("Wikipedia research failed for source discovery, falling back", "topic", opts.TopicName, "error", err)
		}
		if researchCtx != "" {
			prompt = BuildDiscoverPromptWithContext(opts.TopicName, opts.Description, opts.SourcingInstructions, suggested, researchCtx)
		}
	}
	if prompt == "" {
		prompt = BuildDiscoverPrompt(opts.TopicName, opts.Description, opts.SourcingInstructions, suggested)
	}

	resp, err := provider.Chat(ctx, ChatRequest{
		Messages:    []Message{{Role: "user", Content: prompt}},
		Temperature: 0.7,
		MaxTokens:   2048,
		JSONMode:    true,
	})
	if err != nil {
		return nil, 0, provider.Name(), "", err
	}

	responseText := ExtractJSON(resp.Content)
	if responseText == "" {
		return nil, resp.TokensUsed, resp.Provider, resp.Model, fmt.Errorf("empty response from %s", provider.Name())
	}

	var sources []DiscoveredSource
	if err := json.Unmarshal([]byte(responseText), &sources); err != nil {
		return nil, resp.TokensUsed, resp.Provider, resp.Model,
			fmt.Errorf("failed to parse sources JSON from %s: %w (response: %s)", provider.Name(), err, responseText)
	}

	return sources, resp.TokensUsed, resp.Provider, resp.Model, nil
}

// SummarizeContent summarizes scraped content into news stories.
func (c *Client) SummarizeContent(ctx context.Context, opts SummarizeOpts) ([]SummarizedStory, int, string, string, error) {
	if len(opts.ScrapedContent) == 0 {
		return nil, 0, "", "", nil
	}

	provider := c.resolveProvider(opts.AIProvider)

	prompt := BuildSummarizePrompt(
		opts.TopicName, opts.ScrapedContent,
		opts.SummarizingInstructions, opts.ToneInstructions,
		opts.MaxStories, opts.MinWords, opts.MaxWords,
	)

	resp, err := provider.Chat(ctx, ChatRequest{
		Messages:    []Message{{Role: "user", Content: prompt}},
		Temperature: 0.7,
		MaxTokens:   4096,
		JSONMode:    true,
	})
	if err != nil {
		return nil, 0, provider.Name(), "", err
	}

	responseText := ExtractJSON(resp.Content)
	if responseText == "" {
		return nil, resp.TokensUsed, resp.Provider, resp.Model, fmt.Errorf("empty response from %s", provider.Name())
	}

	var stories []SummarizedStory
	if err := json.Unmarshal([]byte(responseText), &stories); err != nil {
		return nil, resp.TokensUsed, resp.Provider, resp.Model,
			fmt.Errorf("failed to parse stories JSON from %s: %w (response: %s)", provider.Name(), err, responseText)
	}

	return stories, resp.TokensUsed, resp.Provider, resp.Model, nil
}

// ListOllamaModels queries the configured Ollama server for available models.
func (c *Client) ListOllamaModels(ctx context.Context) ([]OllamaModel, error) {
	baseURL, _ := c.settings.GetSetting("ollama_url")
	return ListModels(ctx, baseURL)
}

// TestOllamaConnection checks if the configured Ollama server is reachable.
func (c *Client) TestOllamaConnection(ctx context.Context) error {
	baseURL, _ := c.settings.GetSetting("ollama_url")
	return TestConnection(ctx, baseURL)
}

// TestGeminiKey verifies a Gemini API key.
func (c *Client) TestGeminiKey(ctx context.Context, apiKey string) error {
	return c.gemini.TestAPIKey(ctx, apiKey)
}

// GenerateSearchQueries asks the AI to produce search queries for researching a topic.
func (c *Client) GenerateSearchQueries(ctx context.Context, provider Provider, topicName, description string) ([]string, error) {
	prompt := fmt.Sprintf(
		`Generate 3-5 specific search queries for finding factual information about: "%s"
Description: %s

Return ONLY the search queries as a numbered list, one per line. Each query should target a different aspect of the topic.
Make queries specific enough to find Wikipedia articles or authoritative sources.`, topicName, description)

	resp, err := provider.Chat(ctx, ChatRequest{
		Messages:    []Message{{Role: "user", Content: prompt}},
		Temperature: 0.5,
		MaxTokens:   256,
	})
	if err != nil {
		slog.Warn("Failed to generate search queries", "topic", topicName, "error", err)
		// Fallback: use the topic name directly
		return []string{topicName}, nil
	}

	queries := ParseFactsFromText(resp.Content) // reuse numbered-list parser
	if len(queries) == 0 {
		return []string{topicName}, nil
	}
	return queries, nil
}

// ResearchTopic uses AI-generated search queries to find Wikipedia articles,
// then fetches summaries to build a context block for RAG-augmented prompts.
func (c *Client) ResearchTopic(ctx context.Context, provider Provider, topicName, description string) (string, error) {
	if c.wiki == nil {
		return "", fmt.Errorf("wikipedia client not available")
	}

	// Step 1: Ask AI to generate targeted search queries
	queries, err := c.GenerateSearchQueries(ctx, provider, topicName, description)
	if err != nil {
		queries = []string{topicName}
	}

	slog.Debug("Researching niche topic", "topic", topicName, "queries", len(queries))

	// Step 2: Search Wikipedia for each query
	seen := make(map[string]bool)
	var titles []string
	for _, query := range queries {
		results, err := c.wiki.Search(ctx, query, 3)
		if err != nil {
			slog.Debug("Wikipedia search failed", "query", query, "error", err)
			continue
		}
		for _, r := range results {
			if !seen[r.Title] {
				seen[r.Title] = true
				titles = append(titles, r.Title)
			}
		}
	}

	if len(titles) == 0 {
		return "", fmt.Errorf("no Wikipedia articles found for %q", topicName)
	}

	// Step 3: Fetch summaries for top 5 unique articles
	if len(titles) > 5 {
		titles = titles[:5]
	}

	var sb strings.Builder
	for _, title := range titles {
		summary, err := c.wiki.GetSummary(ctx, title)
		if err != nil {
			slog.Debug("Failed to get Wikipedia summary", "title", title, "error", err)
			continue
		}
		sb.WriteString(summary)
		sb.WriteString("\n\n")

		// Cap at ~4000 characters
		if sb.Len() > 4000 {
			break
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("no Wikipedia summaries retrieved for %q", topicName)
	}

	slog.Info("Wikipedia research complete", "topic", topicName, "articles", len(titles), "chars", len(result))
	return result, nil
}
