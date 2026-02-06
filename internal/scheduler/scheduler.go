package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/thinkscotty/kibble/internal/database"
	"github.com/thinkscotty/kibble/internal/gemini"
	"github.com/thinkscotty/kibble/internal/models"
	"github.com/thinkscotty/kibble/internal/similarity"
)

type Scheduler struct {
	db     *database.DB
	gemini *gemini.Client
	sim    *similarity.Checker
	mu     sync.Mutex
}

func New(db *database.DB, gemini *gemini.Client, sim *similarity.Checker) *Scheduler {
	return &Scheduler{db: db, gemini: gemini, sim: sim}
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
