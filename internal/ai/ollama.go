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

// OpenAI-compatible request/response types for Ollama (unexported).

type ollamaChatRequest struct {
	Model          string          `json:"model"`
	Messages       []ollamaMessage `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Stream         bool            `json:"stream"`
	ResponseFormat *ollamaRespFmt  `json:"response_format,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRespFmt struct {
	Type string `json:"type"`
}

type ollamaChatResponse struct {
	Choices []ollamaChoice `json:"choices"`
	Usage   *ollamaUsage   `json:"usage,omitempty"`
	Model   string         `json:"model"`
}

type ollamaChoice struct {
	Message ollamaMessage `json:"message"`
}

type ollamaUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Native Ollama API types for model listing.

type ollamaTagsResponse struct {
	Models []ollamaModelInfo `json:"models"`
}

type ollamaModelInfo struct {
	Name       string             `json:"name"`
	Size       int64              `json:"size"`
	ModifiedAt string             `json:"modified_at"`
	Details    ollamaModelDetails `json:"details"`
}

type ollamaModelDetails struct {
	Family            string `json:"family"`
	ParameterSize     string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

// OllamaProvider implements Provider for Ollama's OpenAI-compatible API.
type OllamaProvider struct {
	httpClient *http.Client
	settings   SettingsGetter
}

// NewOllamaProvider creates an Ollama provider.
func NewOllamaProvider(sg SettingsGetter) *OllamaProvider {
	return &OllamaProvider{
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		settings:   sg,
	}
}

func (o *OllamaProvider) Name() string { return "ollama" }

func (o *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	baseURL, err := o.settings.GetSetting("ollama_url")
	if err != nil || baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model, err := o.settings.GetSetting("ollama_model")
	if err != nil || model == "" {
		model = "mistral-nemo"
	}

	// Convert messages
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

	url := strings.TrimRight(baseURL, "/") + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("parse ollama response: %w", err)
	}

	tokensUsed := 0
	if chatResp.Usage != nil {
		tokensUsed = chatResp.Usage.TotalTokens
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
	}

	return &ChatResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		Model:      model,
		Provider:   "ollama",
	}, nil
}

// ListModels queries the Ollama server for available models.
func ListModels(ctx context.Context, baseURL string) ([]OllamaModel, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/api/tags"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var tagsResp ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	models := make([]OllamaModel, len(tagsResp.Models))
	for i, m := range tagsResp.Models {
		models[i] = OllamaModel{
			Name:          m.Name,
			Size:          m.Size,
			ParameterSize: m.Details.ParameterSize,
			Family:        m.Details.Family,
		}
	}
	return models, nil
}

// TestConnection checks if an Ollama server is reachable.
func TestConnection(ctx context.Context, baseURL string) error {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to Ollama at %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}
	return nil
}
