package ai

// DiscoveredSource is a web source found by AI for a news topic.
type DiscoveredSource struct {
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SummarizedStory is a news story summary produced by AI.
type SummarizedStory struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	SourceURL   string `json:"source_url"`
	SourceTitle string `json:"source_title"`
}

// ScrapedContent holds raw content scraped from a web source.
type ScrapedContent struct {
	URL        string
	SourceName string
	Content    string
}

// OllamaModel represents a model available on an Ollama server.
type OllamaModel struct {
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	ParameterSize string `json:"parameter_size"`
	Family        string `json:"family"`
}

// FactsOpts holds parameters for fact generation.
type FactsOpts struct {
	Topic              string
	Description        string
	CustomInstructions string
	ToneInstructions   string
	Count              int
	MinWords           int
	MaxWords           int
	AIProvider         string // per-topic override: "", "gemini", "ollama"
	IsNiche            bool
}

// DiscoverOpts holds parameters for news source discovery.
type DiscoverOpts struct {
	TopicName            string
	Description          string
	SourcingInstructions string
	AIProvider           string
	IsNiche              bool
	CommunityDomains     []string // Domains frequently shared in related subreddits
}

// SummarizeOpts holds parameters for content summarization.
type SummarizeOpts struct {
	TopicName               string
	ScrapedContent          []ScrapedContent
	SummarizingInstructions string
	ToneInstructions        string
	MaxStories              int
	MinWords                int
	MaxWords                int
	AIProvider              string
	ExistingTitles          []string // Recent story titles for dedup
}
