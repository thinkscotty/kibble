package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/thinkscotty/kibble/internal/database"
	"github.com/thinkscotty/kibble/internal/gemini"
	"github.com/thinkscotty/kibble/internal/models"
	"github.com/thinkscotty/kibble/internal/scraper"
	"github.com/thinkscotty/kibble/internal/similarity"
)

type Scheduler struct {
	db      *database.DB
	gemini  *gemini.Client
	sim     *similarity.Checker
	scraper *scraper.Scraper
	mu      sync.Mutex
}

func New(db *database.DB, gemini *gemini.Client, sim *similarity.Checker, sc *scraper.Scraper) *Scheduler {
	return &Scheduler{db: db, gemini: gemini, sim: sim, scraper: sc}
}

// Run starts the scheduler loop. It checks for due topics every 60 seconds.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	slog.Info("Scheduler started")

	// Run once immediately at startup
	s.checkAndRefresh(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.checkAndRefresh(ctx)
		}
	}
}

func (s *Scheduler) checkAndRefresh(ctx context.Context) {
	// Clean up expired sessions on each tick
	if n, err := s.db.DeleteExpiredSessions(); err != nil {
		slog.Error("Failed to delete expired sessions", "error", err)
	} else if n > 0 {
		slog.Debug("Cleaned up expired sessions", "count", n)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	topics, err := s.db.TopicsDueForRefresh()
	if err != nil {
		slog.Error("Failed to query topics due for refresh", "error", err)
		return
	}

	for _, topic := range topics {
		if ctx.Err() != nil {
			return
		}
		s.refreshTopic(ctx, topic)
	}

	// Check news topics due for refresh
	s.checkAndRefreshNews(ctx)
}

func (s *Scheduler) refreshTopic(ctx context.Context, topic models.Topic) {
	slog.Info("Refreshing topic", "topic", topic.Name, "id", topic.ID)

	customInstr, _ := s.db.GetSetting("ai_custom_instructions")
	toneInstr, _ := s.db.GetSetting("ai_tone_instructions")

	facts, tokensUsed, err := s.gemini.GenerateFacts(
		ctx, topic.Name, topic.Description,
		customInstr, toneInstr, topic.FactsPerRefresh,
	)

	logEntry := models.APIUsageLog{
		TopicID:        &topic.ID,
		FactsRequested: topic.FactsPerRefresh,
		TokensUsed:     tokensUsed,
	}

	if err != nil {
		slog.Error("Failed to generate facts", "topic", topic.Name, "error", err)
		logEntry.ErrorMessage = err.Error()
		s.db.LogAPIUsage(logEntry)
		return
	}

	// Get existing facts for similarity comparison
	existingTrigrams := s.getExistingTrigrams(topic.ID)

	generated := 0
	discarded := 0
	for _, content := range facts {
		if s.sim.IsTooSimilar(content, existingTrigrams) {
			discarded++
			continue
		}

		trigrams := s.sim.Trigrams(content)
		fact := &models.Fact{
			TopicID:  topic.ID,
			Content:  content,
			Trigrams: s.sim.TrigramsToJSON(trigrams),
			Source:   "gemini",
		}
		if err := s.db.CreateFact(fact); err != nil {
			slog.Error("Failed to save fact", "error", err)
			continue
		}

		// Add to existing set so subsequent facts in this batch are also checked
		existingTrigrams = append(existingTrigrams, similarity.StoredTrigrams{
			ID:       fact.ID,
			Trigrams: fact.Trigrams,
		})
		generated++
	}

	logEntry.FactsGenerated = generated
	logEntry.FactsDiscarded = discarded
	s.db.LogAPIUsage(logEntry)
	s.db.UpdateTopicRefreshTime(topic.ID)

	slog.Info("Topic refreshed", "topic", topic.Name,
		"generated", generated, "discarded", discarded)
}

// RefreshNow triggers an immediate refresh for a single topic.
func (s *Scheduler) RefreshNow(ctx context.Context, topicID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, err := s.db.GetTopic(topicID)
	if err != nil {
		return err
	}
	s.refreshTopic(ctx, topic)
	return nil
}

// --- News / Updates scheduling ---

func (s *Scheduler) checkAndRefreshNews(ctx context.Context) {
	newsTopics, err := s.db.NewsTopicsDueForRefresh()
	if err != nil {
		slog.Error("Failed to query news topics due for refresh", "error", err)
		return
	}

	for _, nt := range newsTopics {
		if ctx.Err() != nil {
			return
		}
		s.safeRefreshNewsTopic(ctx, nt.ID)
		// Stagger refreshes to avoid API overload
		time.Sleep(30 * time.Second)
	}
}

func (s *Scheduler) safeRefreshNewsTopic(ctx context.Context, newsTopicID int64) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic in news topic refresh", "topic_id", newsTopicID, "panic", r, "stack", string(debug.Stack()))
			s.db.UpdateNewsRefreshStatus(&models.NewsRefreshStatus{
				NewsTopicID:  newsTopicID,
				NextRefresh:  time.Now().Add(5 * time.Minute),
				Status:       "failed",
				ErrorMessage: fmt.Sprintf("panic: %v", r),
			})
		}
	}()
	s.refreshNewsTopic(ctx, newsTopicID)
}

func (s *Scheduler) refreshNewsTopic(ctx context.Context, newsTopicID int64) {
	topic, err := s.db.GetNewsTopic(newsTopicID)
	if err != nil {
		slog.Error("News topic not found", "id", newsTopicID, "error", err)
		return
	}

	slog.Info("Refreshing news topic", "topic", topic.Name, "id", topic.ID)

	// Mark in-progress
	s.db.UpdateNewsRefreshStatus(&models.NewsRefreshStatus{
		NewsTopicID: newsTopicID,
		Status:      "in_progress",
	})

	// Get active sources
	sources, err := s.db.GetActiveSourcesForNewsTopic(newsTopicID)
	if err != nil {
		s.handleNewsRefreshError(newsTopicID, fmt.Errorf("get sources: %w", err))
		return
	}

	// If no sources, try discovery first
	if len(sources) == 0 {
		if err := s.discoverNewsSources(ctx, newsTopicID); err != nil {
			s.handleNewsRefreshError(newsTopicID, fmt.Errorf("discover sources: %w", err))
			return
		}
		sources, _ = s.db.GetActiveSourcesForNewsTopic(newsTopicID)
		if len(sources) == 0 {
			s.handleNewsRefreshError(newsTopicID, fmt.Errorf("no sources available for topic"))
			return
		}
	}

	// Scrape content
	scrapeCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	scrapeResults := s.scraper.ScrapeSources(scrapeCtx, sources)

	// Process results and update source statuses
	var scrapedContent []gemini.ScrapedContent
	for _, result := range scrapeResults {
		if result.Error != nil {
			newFailureCount := result.Source.FailureCount + 1
			isActive := newFailureCount < 3

			errMsg := result.Error.Error()
			if len(errMsg) > 500 {
				errMsg = errMsg[:500]
			}

			s.db.UpdateNewsSourceStatus(result.Source.ID, isActive, newFailureCount, errMsg)
			if !isActive {
				slog.Warn("News source disabled after failures", "url", result.Source.URL, "failures", newFailureCount)
			}
		} else {
			// Reset failure count on success
			if result.Source.FailureCount > 0 {
				s.db.UpdateNewsSourceStatus(result.Source.ID, true, 0, "")
			}
			scrapedContent = append(scrapedContent, *result.Content)
		}
	}

	if len(scrapedContent) == 0 {
		s.handleNewsRefreshError(newsTopicID, fmt.Errorf("failed to scrape any content from active sources"))
		return
	}

	// Summarize with Gemini
	summarizeInstr, _ := s.db.GetSetting("news_summarizing_instructions")
	toneInstr, _ := s.db.GetSetting("news_tone_instructions")

	stories, _, err := s.gemini.SummarizeContent(
		ctx, topic.Name, scrapedContent,
		summarizeInstr, toneInstr, topic.StoriesPerRefresh,
	)
	if err != nil {
		s.handleNewsRefreshError(newsTopicID, fmt.Errorf("summarize content: %w", err))
		return
	}

	// Store stories
	for _, story := range stories {
		dbStory := &models.Story{
			NewsTopicID: newsTopicID,
			Title:       story.Title,
			Summary:     story.Summary,
			SourceURL:   story.SourceURL,
			SourceTitle: story.SourceTitle,
		}
		if err := s.db.CreateStory(dbStory); err != nil {
			slog.Error("Failed to create story", "error", err)
		}
	}

	// Clean up old stories (keep 3x display count)
	s.db.DeleteOldStories(newsTopicID, topic.StoriesPerRefresh*3)

	// Mark completed
	s.db.UpdateNewsRefreshStatus(&models.NewsRefreshStatus{
		NewsTopicID: newsTopicID,
		LastRefresh: time.Now(),
		NextRefresh: time.Now().Add(time.Duration(topic.RefreshIntervalMinutes) * time.Minute),
		Status:      "completed",
	})
	s.db.UpdateNewsTopicRefreshTime(newsTopicID)

	slog.Info("News topic refreshed", "topic", topic.Name, "stories", len(stories))
}

func (s *Scheduler) discoverNewsSources(ctx context.Context, newsTopicID int64) error {
	topic, err := s.db.GetNewsTopic(newsTopicID)
	if err != nil {
		return fmt.Errorf("topic not found: %w", err)
	}

	sourcingInstr, _ := s.db.GetSetting("news_sourcing_instructions")

	sources, _, err := s.gemini.DiscoverSources(ctx, topic.Name, topic.Description, sourcingInstr)
	if err != nil {
		return fmt.Errorf("discover sources: %w", err)
	}

	// Clear existing AI sources and add new ones
	s.db.ClearAINewsSourcesForTopic(newsTopicID)

	for _, source := range sources {
		if err := scraper.ValidateURL(source.URL); err != nil {
			slog.Debug("Skipping invalid source URL", "url", source.URL, "error", err)
			continue
		}
		if _, err := s.db.AddNewsSource(newsTopicID, source.URL, source.Name, false); err != nil {
			slog.Error("Failed to add news source", "error", err)
		}
	}

	slog.Info("Discovered news sources", "topic", topic.Name, "count", len(sources))
	return nil
}

func (s *Scheduler) handleNewsRefreshError(newsTopicID int64, err error) {
	slog.Error("News refresh error", "topic_id", newsTopicID, "error", err)
	s.db.UpdateNewsRefreshStatus(&models.NewsRefreshStatus{
		NewsTopicID:  newsTopicID,
		NextRefresh:  time.Now().Add(5 * time.Minute),
		Status:       "failed",
		ErrorMessage: err.Error(),
	})
}

// RefreshNewsNow triggers an immediate news topic refresh.
func (s *Scheduler) RefreshNewsNow(ctx context.Context, newsTopicID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.safeRefreshNewsTopic(ctx, newsTopicID)
}

// DiscoverSourcesNow triggers immediate source discovery for a news topic.
func (s *Scheduler) DiscoverSourcesNow(ctx context.Context, newsTopicID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.discoverNewsSources(ctx, newsTopicID)
}

func (s *Scheduler) getExistingTrigrams(topicID int64) []similarity.StoredTrigrams {
	dbTrigrams, err := s.db.GetFactTrigramsForTopic(topicID)
	if err != nil {
		slog.Error("Failed to get existing trigrams", "error", err)
		return nil
	}

	result := make([]similarity.StoredTrigrams, len(dbTrigrams))
	for i, dt := range dbTrigrams {
		result[i] = similarity.StoredTrigrams{
			ID:       dt.ID,
			Trigrams: dt.Trigrams,
		}
	}
	return result
}
