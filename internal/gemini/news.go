package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/thinkscotty/kibble/internal/feeds"
)

// DiscoverSources uses AI to find relevant web sources for a news topic.
func (c *Client) DiscoverSources(ctx context.Context, topicName, description, sourcingInstructions string) ([]DiscoveredSource, int, error) {
	apiKey, err := c.db.GetSetting("gemini_api_key")
	if err != nil || apiKey == "" {
		return nil, 0, fmt.Errorf("gemini API key not configured — set it in Settings")
	}

	suggested := feeds.FindRelevant(topicName, description)
	prompt := buildDiscoverPrompt(topicName, description, sourcingInstructions, suggested)

	reqBody := GenerateRequest{
		Contents: []Content{{
			Parts: []Part{{Text: prompt}},
		}},
		GenerationConfig: &GenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 2048,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	url := baseURL + "?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(body))
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, 0, fmt.Errorf("parse gemini response: %w", err)
	}

	tokensUsed := 0
	if genResp.UsageMetadata != nil {
		tokensUsed = genResp.UsageMetadata.TotalTokenCount
	}

	responseText := extractResponseText(genResp)
	if responseText == "" {
		return nil, tokensUsed, fmt.Errorf("empty response from Gemini")
	}

	responseText = cleanJSONResponse(responseText)

	var sources []DiscoveredSource
	if err := json.Unmarshal([]byte(responseText), &sources); err != nil {
		return nil, tokensUsed, fmt.Errorf("failed to parse sources JSON: %w (response: %s)", err, responseText)
	}

	return sources, tokensUsed, nil
}

// SummarizeContent summarizes scraped content into news stories.
func (c *Client) SummarizeContent(ctx context.Context, topicName string, scrapedContent []ScrapedContent, summarizingInstructions, toneInstructions string, maxStories, minWords, maxWords int) ([]SummarizedStory, int, error) {
	if len(scrapedContent) == 0 {
		return nil, 0, nil
	}

	apiKey, err := c.db.GetSetting("gemini_api_key")
	if err != nil || apiKey == "" {
		return nil, 0, fmt.Errorf("gemini API key not configured — set it in Settings")
	}

	prompt := buildSummarizePrompt(topicName, scrapedContent, summarizingInstructions, toneInstructions, maxStories, minWords, maxWords)

	reqBody := GenerateRequest{
		Contents: []Content{{
			Parts: []Part{{Text: prompt}},
		}},
		GenerationConfig: &GenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 4096,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	url := baseURL + "?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(body))
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, 0, fmt.Errorf("parse gemini response: %w", err)
	}

	tokensUsed := 0
	if genResp.UsageMetadata != nil {
		tokensUsed = genResp.UsageMetadata.TotalTokenCount
	}

	responseText := extractResponseText(genResp)
	if responseText == "" {
		return nil, tokensUsed, fmt.Errorf("empty response from Gemini")
	}

	responseText = cleanJSONResponse(responseText)

	var stories []SummarizedStory
	if err := json.Unmarshal([]byte(responseText), &stories); err != nil {
		return nil, tokensUsed, fmt.Errorf("failed to parse stories JSON: %w (response: %s)", err, responseText)
	}

	return stories, tokensUsed, nil
}

func buildDiscoverPrompt(topicName, description, sourcingInstructions string, suggestedFeeds []feeds.Feed) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are a helpful assistant that discovers reliable web sources for news topics.

Topic: %s
Description: %s

`, topicName, description))

	if sourcingInstructions != "" {
		sb.WriteString(sourcingInstructions)
		sb.WriteString("\n\n")
	}

	if len(suggestedFeeds) > 0 {
		sb.WriteString("Here are known-good RSS feeds that may be relevant to this topic. PREFER these feeds when they match the topic well, as they are verified to work:\n\n")
		for _, f := range suggestedFeeds {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", f.Name, f.URL))
		}
		sb.WriteString("\nYou may include additional sources beyond this list if needed to cover the topic well.\n\n")
	}

	sb.WriteString(`Find 4-8 reliable sources that provide ongoing news and updates related to this topic. Sources can include:
- News websites and RSS feeds
- Reddit subreddits (format as https://reddit.com/r/subredditname)
- Technical blogs or official sources

For Reddit, include 1-2 relevant subreddits if they exist for this topic. Choose active subreddits with engaged communities.

For each source, provide:
1. The URL (must be a real, working URL)
2. A short name for the source
3. A brief description of what content it provides

IMPORTANT: Return ONLY a valid JSON array with no additional text, markdown, or explanation.

Format:
[
  {"url": "https://example.com/feed", "name": "Example News", "description": "Daily updates on topic"},
  {"url": "https://reddit.com/r/technology", "name": "r/technology", "description": "Tech news and discussion"}
]`)

	return sb.String()
}

func buildSummarizePrompt(topicName string, scrapedContent []ScrapedContent, summarizingInstructions, toneInstructions string, maxStories, minWords, maxWords int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are a news summarization assistant. Analyze the following scraped content and create clear, informative news summaries.

Topic: %s

`, topicName))

	if summarizingInstructions != "" {
		sb.WriteString(summarizingInstructions)
		sb.WriteString("\n\n")
	}

	if toneInstructions != "" {
		sb.WriteString("Tone and style: ")
		sb.WriteString(toneInstructions)
		sb.WriteString("\n\n")
	}

	if minWords > 0 && maxWords > 0 {
		sb.WriteString(fmt.Sprintf("Each story summary should be between %d and %d words long.\n\n", minWords, maxWords))
	} else if minWords > 0 {
		sb.WriteString(fmt.Sprintf("Each story summary should be at least %d words long.\n\n", minWords))
	} else if maxWords > 0 {
		sb.WriteString(fmt.Sprintf("Each story summary should be at most %d words long.\n\n", maxWords))
	}

	sb.WriteString("Scraped Content:\n")
	for i, content := range scrapedContent {
		sb.WriteString(fmt.Sprintf("\n--- Source %d: %s ---\nURL: %s\n%s\n",
			i+1, content.SourceName, content.URL, content.Content))
	}

	sb.WriteString(fmt.Sprintf(`
From the content above, identify the %d most interesting and relevant news stories.

IMPORTANT FILTERING RULES:
- ONLY include content that DIRECTLY relates to the topic "%s"
- Skip any content that is off-topic or only tangentially related
- For Reddit posts, focus on substantive discussions and news, not casual comments or memes
- Prioritize recent, newsworthy content over general discussion

For each story:
1. Create a compelling headline (title)
2. Write a summary focusing on key facts and why this story matters
3. Include the source URL where the story was found
4. Include the source name/title

IMPORTANT: Return ONLY a valid JSON array with no additional text, markdown, or explanation.

Format:
[
  {"title": "Headline Here", "summary": "Summary text here...", "source_url": "https://source.com/article", "source_title": "Source Name"}
]`, maxStories, topicName))

	return sb.String()
}

func extractResponseText(resp GenerateResponse) string {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	return resp.Candidates[0].Content.Parts[0].Text
}

func cleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}
	if strings.HasSuffix(response, "```") {
		response = strings.TrimSuffix(response, "```")
	}
	return strings.TrimSpace(response)
}
