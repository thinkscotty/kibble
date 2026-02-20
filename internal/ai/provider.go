package ai

import "context"

// SettingsGetter is a minimal interface so the ai package does not import database.
type SettingsGetter interface {
	GetSetting(key string) (string, error)
}

// Provider is the interface that all AI backends must implement.
type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	Name() string // "gemini", "ollama", or "chutes"
}

// ChatRequest is a provider-agnostic request.
type ChatRequest struct {
	Messages    []Message
	Temperature float64
	MaxTokens   int
	JSONMode    bool // request JSON-formatted output
}

// ChatResponse is a provider-agnostic response.
type ChatResponse struct {
	Content    string
	TokensUsed int
	Model      string // e.g. "gemini-2.5-flash" or "mistral-nemo"
	Provider   string // "gemini" or "ollama"
}

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}
