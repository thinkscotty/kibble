package gemini

// --- Request types for Gemini REST API ---

type GenerateRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

// --- Response types ---

type GenerateResponse struct {
	Candidates    []Candidate    `json:"candidates"`
	UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// --- News / Updates types ---

type DiscoveredSource struct {
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SummarizedStory struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	SourceURL   string `json:"source_url"`
	SourceTitle string `json:"source_title"`
}

type ScrapedContent struct {
	URL        string
	SourceName string
	Content    string
}
