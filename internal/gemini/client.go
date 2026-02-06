package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

// SettingsGetter is a minimal interface so gemini package does not import database.
type SettingsGetter interface {
	GetSetting(key string) (string, error)
}

type Client struct {
	httpClient *http.Client
	db         SettingsGetter
}

func NewClient(sg SettingsGetter) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		db:         sg,
	}
}

// GenerateFacts calls the Gemini API and returns parsed facts and token count.
func (c *Client) GenerateFacts(ctx context.Context, topic, description, customInstructions, toneInstructions string, count int) ([]string, int, error) {
	apiKey, err := c.db.GetSetting("gemini_api_key")
	if err != nil || apiKey == "" {
		return nil, 0, fmt.Errorf("gemini API key not configured — set it in Settings")
	}

	prompt := BuildPrompt(topic, description, customInstructions, toneInstructions, count)

	reqBody := GenerateRequest{
		Contents: []Content{{
			Parts: []Part{{Text: prompt}},
		}},
		GenerationConfig: &GenerationConfig{
			Temperature:     0.9,
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

	facts := ParseFacts(genResp)
	return facts, tokensUsed, nil
}

// TestAPIKey verifies the API key works by sending a minimal request.
func (c *Client) TestAPIKey(ctx context.Context, apiKey string) error {
	reqBody := GenerateRequest{
		Contents: []Content{{
			Parts: []Part{{Text: "Say hello in one word."}},
		}},
		GenerationConfig: &GenerationConfig{
			MaxOutputTokens: 10,
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	url := baseURL + "?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
