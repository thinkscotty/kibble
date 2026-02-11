package models

import "time"

type Topic struct {
	ID                     int64      `json:"id"`
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	DisplayOrder           int        `json:"display_order"`
	IsActive               bool       `json:"is_active"`
	FactsPerRefresh        int        `json:"facts_per_refresh"`
	RefreshIntervalMinutes int        `json:"refresh_interval_minutes"`
	LastRefreshedAt        *time.Time `json:"last_refreshed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type Fact struct {
	ID         int64     `json:"id"`
	TopicID    int64     `json:"topic_id"`
	TopicName  string    `json:"topic_name,omitempty"`
	Content    string    `json:"content"`
	Trigrams   string    `json:"-"`
	IsCustom   bool      `json:"is_custom"`
	IsArchived bool      `json:"is_archived"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type APIUsageLog struct {
	ID             int64     `json:"id"`
	TopicID        *int64    `json:"topic_id,omitempty"`
	TopicName      string    `json:"topic_name,omitempty"`
	FactsRequested int       `json:"facts_requested"`
	FactsGenerated int       `json:"facts_generated"`
	FactsDiscarded int       `json:"facts_discarded"`
	TokensUsed     int       `json:"tokens_used"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	Token     string    `json:"-"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type TopicWithFacts struct {
	Topic Topic
	Facts []Fact
}

type NewsTopic struct {
	ID                     int64      `json:"id"`
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	DisplayOrder           int        `json:"display_order"`
	IsActive               bool       `json:"is_active"`
	StoriesPerRefresh      int        `json:"stories_per_refresh"`
	RefreshIntervalMinutes int        `json:"refresh_interval_minutes"`
	LastRefreshedAt        *time.Time `json:"last_refreshed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type NewsSource struct {
	ID           int64     `json:"id"`
	NewsTopicID  int64     `json:"news_topic_id"`
	URL          string    `json:"url"`
	Name         string    `json:"name"`
	IsManual     bool      `json:"is_manual"`
	IsActive     bool      `json:"is_active"`
	FailureCount int       `json:"failure_count"`
	LastError    string    `json:"last_error"`
	CreatedAt    time.Time `json:"created_at"`
}

type Story struct {
	ID          int64     `json:"id"`
	NewsTopicID int64     `json:"news_topic_id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	SourceURL   string    `json:"source_url"`
	SourceTitle string    `json:"source_title"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type NewsTopicWithStories struct {
	NewsTopic NewsTopic
	Stories   []Story
}

type NewsTopicWithSources struct {
	NewsTopic NewsTopic
	Sources   []NewsSource
}

type NewsRefreshStatus struct {
	NewsTopicID  int64     `json:"news_topic_id"`
	LastRefresh  time.Time `json:"last_refresh"`
	NextRefresh  time.Time `json:"next_refresh"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

type Stats struct {
	TotalTopics      int   `json:"total_topics"`
	ActiveTopics     int   `json:"active_topics"`
	TotalFacts       int   `json:"total_facts"`
	CustomFacts      int   `json:"custom_facts"`
	AIGeneratedFacts int   `json:"ai_generated_facts"`
	TotalAPIRequests int   `json:"total_api_requests"`
	TotalTokensUsed  int   `json:"total_tokens_used"`
	FactsDiscarded   int   `json:"facts_discarded"`
	DatabaseSizeBytes int64 `json:"database_size_bytes"`
}
