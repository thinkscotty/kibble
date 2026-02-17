package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

// Gemini API request/response types (unexported).

type geminiRequest struct {
	Contents         []geminiContent    `json:"contents"`
	GenerationConfig *geminiGenConfig   `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate  `json:"candidates"`
	UsageMetadata *geminiUsage       `json:"usageMetadata,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GeminiProvider implements Provider for Google's Gemini API.
type GeminiProvider struct {
	httpClient *http.Client
	settings   SettingsGetter
}

// NewGeminiProvider creates a Gemini provider.
func NewGeminiProvider(sg SettingsGetter) *GeminiProvider {
	return &GeminiProvider{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		settings:   sg,
	}
}

func (g *GeminiProvider) Name() string { return "gemini" }

func (g *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	apiKey, err := g.settings.GetSetting("gemini_api_key")
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("gemini API key not configured — set it in Settings")
	}

	// Build the prompt from messages (Gemini uses a single content block)
	prompt := messagesToPrompt(req.Messages)

	body := geminiRequest{
		Contents: []geminiContent{{
			Parts: []geminiPart{{Text: prompt}},
		}},
		GenerationConfig: &geminiGenConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		},
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := geminiBaseURL + "?key=" + apiKey
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var genResp geminiResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return nil, fmt.Errorf("parse gemini response: %w", err)
	}

	tokensUsed := 0
	if genResp.UsageMetadata != nil {
		tokensUsed = genResp.UsageMetadata.TotalTokenCount
	}

	content := ""
	if len(genResp.Candidates) > 0 && len(genResp.Candidates[0].Content.Parts) > 0 {
		content = genResp.Candidates[0].Content.Parts[0].Text
	}

	return &ChatResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      "gemini-2.5-flash",
		Provider:   "gemini",
	}, nil
}

// TestAPIKey verifies a Gemini API key by sending a minimal request.
func (g *GeminiProvider) TestAPIKey(ctx context.Context, apiKey string) error {
	body := geminiRequest{
		Contents: []geminiContent{{
			Parts: []geminiPart{{Text: "Say hello in one word."}},
		}},
		GenerationConfig: &geminiGenConfig{
			MaxOutputTokens: 10,
		},
	}

	jsonData, _ := json.Marshal(body)
	url := geminiBaseURL + "?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// messagesToPrompt concatenates chat messages into a single prompt string for Gemini.
// Gemini's simple API doesn't support role-based messages, so we format them as text.
func messagesToPrompt(messages []Message) string {
	if len(messages) == 1 {
		return messages[0].Content
	}

	var sb strings.Builder
	for _, m := range messages {
		if m.Role == "system" {
			sb.WriteString(m.Content)
			sb.WriteString("\n\n")
		} else {
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
