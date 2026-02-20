package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const chutesBaseURL = "https://llm.chutes.ai/v1/chat/completions"

// ChutesProvider implements Provider for the Chutes.ai OpenAI-compatible API.
type ChutesProvider struct {
	httpClient *http.Client
	settings   SettingsGetter
}

// NewChutesProvider creates a Chutes.ai provider.
func NewChutesProvider(sg SettingsGetter) *ChutesProvider {
	return &ChutesProvider{
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		settings:   sg,
	}
}

func (c *ChutesProvider) Name() string { return "chutes" }

func (c *ChutesProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	apiKey, err := c.settings.GetSetting("chutes_api_key")
	if err != nil || strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("chutes API key not configured — set it in Settings")
	}
	apiKey = strings.TrimSpace(apiKey)

	model, err := c.settings.GetSetting("chutes_model")
	if err != nil || strings.TrimSpace(model) == "" {
		model = "deepseek-ai/DeepSeek-V3"
	}
	model = strings.TrimSpace(model)

	if ctx.Err() != nil {
		return nil, fmt.Errorf("chutes request skipped (context already cancelled): %w", ctx.Err())
	}

	// Reuse OpenAI-compatible types from ollama.go (same package, same format)
	msgs := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = ollamaMessage{Role: m.Role, Content: m.Content}
	}

	body := ollamaChatRequest{
		Model:       model,
		Messages:    msgs,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	if req.JSONMode {
		body.ResponseFormat = &ollamaRespFmt{Type: "json_object"}
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	promptChars := 0
	for _, m := range msgs {
		promptChars += len(m.Content)
	}

	slog.Info("Chutes request starting", "model", model, "prompt_chars", promptChars, "json_mode", req.JSONMode)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chutesBaseURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		slog.Error("Chutes request failed", "model", model, "elapsed", time.Since(start), "error", err)
		return nil, fmt.Errorf("chutes request failed (model=%s, elapsed=%s): %w", model, time.Since(start), err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		errMsg := extractOllamaError(respBody) // works for any OpenAI-compatible error format
		if errMsg == "" {
			errMsg = string(respBody)
		}
		slog.Error("Chutes API error", "status", resp.StatusCode, "model", model, "error", errMsg)
		return nil, fmt.Errorf("chutes returned status %d: %s", resp.StatusCode, errMsg)
	}

	var chatResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("parse chutes response: %w", err)
	}

	tokensUsed := 0
	if chatResp.Usage != nil {
		tokensUsed = chatResp.Usage.TotalTokens
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
	}

	slog.Info("Chutes request completed", "model", model, "elapsed", time.Since(start), "tokens", tokensUsed, "response_chars", len(content))

	return &ChatResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      model,
		Provider:   "chutes",
	}, nil
}

// TestChutesKey verifies a Chutes.ai API key by sending a minimal request.
func TestChutesKey(ctx context.Context, apiKey, model string) error {
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("API key is empty")
	}
	if strings.TrimSpace(model) == "" {
		model = "deepseek-ai/DeepSeek-V3"
	}

	body := ollamaChatRequest{
		Model:     model,
		Messages:  []ollamaMessage{{Role: "user", Content: "Say hello in exactly one word."}},
		MaxTokens: 10,
		Stream:    false,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chutesBaseURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key (401 Unauthorized)")
	}
	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		errMsg := extractOllamaError(respBody)
		if errMsg == "" {
			errMsg = string(respBody)
		}
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, errMsg)
	}

	return nil
}
