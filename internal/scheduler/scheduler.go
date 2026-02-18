package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/thinkscotty/kibble/internal/ai"
	"github.com/thinkscotty/kibble/internal/database"
	"github.com/thinkscotty/kibble/internal/feeds"
	"github.com/thinkscotty/kibble/internal/models"
	"github.com/thinkscotty/kibble/internal/reddit"
	"github.com/thinkscotty/kibble/internal/scraper"
	"github.com/thinkscotty/kibble/internal/similarity"
)

type Scheduler struct {
	db      *database.DB
	ai      *ai.Client
	sim     *similarity.Checker
	scraper *scraper.Scraper
	locks   sync.Map // per-topic locks: topicKey -> *sync.Mutex
}

// topicKey returns a unique key for per-topic locking.
func topicKey(kind string, id int64) string {
	return fmt.Sprintf("%s:%d", kind, id)
}

// lockTopic acquires a per-topic mutex, creating it if needed.
// Returns the mutex (caller must Unlock) and true if the lock was acquired.
// Returns nil and false if the topic is already locked (non-blocking).
func (s *Scheduler) lockTopic(key string) (*sync.Mutex, bool) {
	val, _ := s.locks.LoadOrStore(key, &sync.Mutex{})
	mu := val.(*sync.Mutex)
	if mu.TryLock() {
		return mu, true
	}
	return nil, false
}

func New(db *database.DB, aiClient *ai.Client, sim *similarity.Checker, sc *scraper.Scraper) *Scheduler {
	return &Scheduler{db: db, ai: aiClient, sim: sim, scraper: sc}
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

	// Refresh fact topics concurrently (up to 3 at a time)
	topics, err := s.db.TopicsDueForRefresh()
	if err != nil {
		slog.Error("Failed to query topics due for refresh", "error", err)
	} else if len(topics) > 0 {
		sem := make(chan struct{}, 3)
		var wg sync.WaitGroup
		for _, topic := range topics {
			if ctx.Err() != nil {
				break
			}
			wg.Add(1)
			go func(t models.Topic) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				key := topicKey("fact", t.ID)
				mu, ok := s.lockTopic(key)
				if !ok {
					slog.Debug("Topic already being refreshed, skipping", "topic", t.Name)
					return
				}
				defer mu.Unlock()
				s.refreshTopic(ctx, t)
			}(topic)
		}
		wg.Wait()
	}

	// Refresh news topics concurrently (up to 2 at a time)
	s.checkAndRefreshNews(ctx)
}

func (s *Scheduler) refreshTopic(ctx context.Context, topic models.Topic) {
	slog.Info("Refreshing topic", "topic", topic.Name, "id", topic.ID)
	start := time.Now()

	customInstr, _ := s.db.GetSetting("ai_custom_instructions")
	toneInstr, _ := s.db.GetSetting("ai_tone_instructions")

	aiCtx, aiCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer aiCancel()

	facts, tokensUsed, providerName, modelName, err := s.ai.GenerateFacts(aiCtx, ai.FactsOpts{
		Topic:              topic.Name,
		Description:        topic.Description,
		CustomInstructions: customInstr,
		ToneInstructions:   toneInstr,
		Count:              topic.FactsPerRefresh,
		MinWords:           topic.SummaryMinWords,
		MaxWords:           topic.SummaryMaxWords,
		AIProvider:         topic.AIProvider,
		IsNiche:            topic.IsNiche,
	})

	logEntry := models.APIUsageLog{
		TopicID:        &topic.ID,
		FactsRequested: topic.FactsPerRefresh,
		TokensUsed:     tokensUsed,
		AIProvider:     providerName,
		AIModel:        modelName,
	}

	if err != nil {
		slog.Error("Failed to generate facts", "topic", topic.Name, "error", err)
		logEntry.ErrorMessage = err.Error()
		s.db.LogAPIUsage(logEntry)
		s.db.LogRefresh(models.RefreshLog{
			TopicType: "facts", TopicID: topic.ID, TopicName: topic.Name,
			Status: "error", ErrorType: classifyError(err), ErrorMessage: err.Error(),
			DurationMs: time.Since(start).Milliseconds(),
			AIProvider: providerName, AIModel: modelName,
		})
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
			TopicID:    topic.ID,
			Content:    content,
			Trigrams:   s.sim.TrigramsToJSON(trigrams),
			Source:     providerName,
			AIProvider: providerName,
			AIModel:    modelName,
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

	s.db.LogRefresh(models.RefreshLog{
		TopicType: "facts", TopicID: topic.ID, TopicName: topic.Name,
		Status: "success", DurationMs: time.Since(start).Milliseconds(),
		AIProvider: providerName, AIModel: modelName, ItemCount: generated,
	})

	slog.Info("Topic refreshed", "topic", topic.Name,
		"generated", generated, "discarded", discarded)
}

// RefreshNow triggers an immediate refresh for a single topic.
func (s *Scheduler) RefreshNow(ctx context.Context, topicID int64) error {
	key := topicKey("fact", topicID)
	mu, ok := s.lockTopic(key)
	if !ok {
		return fmt.Errorf("topic is already being refreshed")
	}
	defer mu.Unlock()

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

	if len(newsTopics) == 0 {
		return
	}

	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup
	for _, nt := range newsTopics {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			key := topicKey("news", id)
			mu, ok := s.lockTopic(key)
			if !ok {
				slog.Debug("News topic already being refreshed, skipping", "topic_id", id)
				return
			}
			defer mu.Unlock()
			s.safeRefreshNewsTopic(ctx, id)
		}(nt.ID)
	}
	wg.Wait()
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
	start := time.Now()

	// Mark in-progress
	s.db.UpdateNewsRefreshStatus(&models.NewsRefreshStatus{
		NewsTopicID: newsTopicID,
		Status:      "in_progress",
	})

	// Get active sources
	sources, err := s.db.GetActiveSourcesForNewsTopic(newsTopicID)
	if err != nil {
		s.handleNewsRefreshError(newsTopicID, fmt.Errorf("get sources: %w", err))
		s.logNewsRefreshError(topic, start, fmt.Errorf("get sources: %w", err))
		return
	}

	// If no sources, try discovery first
	if len(sources) == 0 {
		if err := s.discoverNewsSources(ctx, newsTopicID); err != nil {
			s.handleNewsRefreshError(newsTopicID, fmt.Errorf("discover sources: %w", err))
			s.logNewsRefreshError(topic, start, fmt.Errorf("discover sources: %w", err))
			return
		}
		sources, _ = s.db.GetActiveSourcesForNewsTopic(newsTopicID)
		if len(sources) == 0 {
			noSourcesErr := fmt.Errorf("no sources available for topic")
			s.handleNewsRefreshError(newsTopicID, noSourcesErr)
			s.logNewsRefreshError(topic, start, noSourcesErr)
			return
		}
	}

	// Scrape content
	scrapeCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	scrapeResults := s.scraper.ScrapeSources(scrapeCtx, sources)

	// Process results and update source statuses.
	// Failure count increments on each failed refresh and decrements by 1
	// on each success (min 0). This lets occasional failures be forgiven
	// while chronically bad sources accumulate toward the removal threshold.
	var scrapedContent []ai.ScrapedContent
	var removedSourceCount int
	for _, result := range scrapeResults {
		if result.Error != nil {
			newFailureCount := result.Source.FailureCount + 1

			errMsg := result.Error.Error()
			if len(errMsg) > 500 {
				errMsg = errMsg[:500]
			}

			if newFailureCount >= 3 {
				// Auto-remove source after accumulating 3 failures across refreshes
				s.db.DeleteNewsSource(result.Source.ID)
				removedSourceCount++
				slog.Warn("Auto-removed failing news source",
					"url", result.Source.URL, "name", result.Source.Name,
					"failures", newFailureCount, "topic_id", newsTopicID)
			} else {
				s.db.UpdateNewsSourceStatus(result.Source.ID, true, newFailureCount, errMsg)
			}
		} else {
			// Decrement failure count by 1 on success (min 0) so that
			// occasional failures are forgiven over time
			if result.Source.FailureCount > 0 {
				s.db.UpdateNewsSourceStatus(result.Source.ID, true, result.Source.FailureCount-1, "")
			}
			scrapedContent = append(scrapedContent, *result.Content)
		}
	}

	// Discover replacement sources for any that were auto-removed
	if removedSourceCount > 0 {
		s.replaceRemovedSources(ctx, newsTopicID, removedSourceCount)
	}

	if len(scrapedContent) == 0 {
		noContentErr := fmt.Errorf("failed to scrape any content from active sources")
		s.handleNewsRefreshError(newsTopicID, noContentErr)
		s.logNewsRefreshError(topic, start, noContentErr)
		return
	}

	// Summarize with AI
	summarizeInstr, _ := s.db.GetSetting("news_summarizing_instructions")
	toneInstr, _ := s.db.GetSetting("news_tone_instructions")

	// Fetch recent story titles for deduplication context
	existingTitles, _ := s.db.GetRecentStoryTitles(newsTopicID, 30)

	sumCtx, sumCancel := context.WithTimeout(ctx, 8*time.Minute)
	defer sumCancel()

	stories, _, storyProvider, storyModel, err := s.ai.SummarizeContent(sumCtx, ai.SummarizeOpts{
		TopicName:               topic.Name,
		ScrapedContent:          scrapedContent,
		SummarizingInstructions: summarizeInstr,
		ToneInstructions:        toneInstr,
		MaxStories:              topic.StoriesPerRefresh,
		MinWords:                topic.SummaryMinWords,
		MaxWords:                topic.SummaryMaxWords,
		AIProvider:              topic.AIProvider,
		ExistingTitles:          existingTitles,
	})
	if err != nil {
		s.handleNewsRefreshError(newsTopicID, fmt.Errorf("summarize content: %w", err))
		s.logNewsRefreshError(topic, start, fmt.Errorf("summarize content: %w", err))
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
			AIProvider:  storyProvider,
			AIModel:     storyModel,
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

	s.db.LogRefresh(models.RefreshLog{
		TopicType: "news", TopicID: topic.ID, TopicName: topic.Name,
		Status: "success", DurationMs: time.Since(start).Milliseconds(),
		AIProvider: storyProvider, AIModel: storyModel, ItemCount: len(stories),
	})

	slog.Info("News topic refreshed", "topic", topic.Name, "stories", len(stories))
}

func (s *Scheduler) discoverNewsSources(ctx context.Context, newsTopicID int64) error {
	topic, err := s.db.GetNewsTopic(newsTopicID)
	if err != nil {
		return fmt.Errorf("topic not found: %w", err)
	}

	sourcingInstr, _ := s.db.GetSetting("news_sourcing_instructions")

	// Mine Reddit subreddits for frequently-shared external sources
	communityDomains := s.mineRedditDomains(ctx, newsTopicID, topic.Name, topic.Description)

	discoverCtx, discoverCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer discoverCancel()

	sources, _, _, _, err := s.ai.DiscoverSources(discoverCtx, ai.DiscoverOpts{
		TopicName:            topic.Name,
		Description:          topic.Description,
		SourcingInstructions: sourcingInstr,
		AIProvider:           topic.AIProvider,
		IsNiche:              topic.IsNiche,
		CommunityDomains:     communityDomains,
	})
	if err != nil {
		return fmt.Errorf("discover sources: %w", err)
	}

	// Clear existing AI sources and add new ones
	s.db.ClearAINewsSourcesForTopic(newsTopicID)

	var accepted int
	for _, source := range sources {
		if err := scraper.ValidateURL(source.URL); err != nil {
			slog.Debug("Skipping invalid source URL", "url", source.URL, "error", err)
			continue
		}

		// Validate source: test-scrape + RSS auto-discovery
		result := s.scraper.ValidateSource(ctx, source.URL, source.Name)
		if !result.OK {
			slog.Info("Rejected news source (validation failed)",
				"url", source.URL, "name", source.Name, "reason", result.Reason)
			continue
		}

		finalURL := source.URL
		if result.FeedURL != "" {
			slog.Info("Discovered RSS feed for source", "original", source.URL, "rss", result.FeedURL)
			finalURL = result.FeedURL
		}

		if _, err := s.db.AddNewsSource(newsTopicID, finalURL, source.Name, false); err != nil {
			slog.Error("Failed to add news source", "error", err)
			continue
		}
		accepted++
	}

	slog.Info("Discovered news sources", "topic", topic.Name, "discovered", len(sources), "accepted", accepted)
	return nil
}

// replaceRemovedSources discovers new sources to replace ones that were auto-removed due to failures.
func (s *Scheduler) replaceRemovedSources(ctx context.Context, newsTopicID int64, count int) {
	topic, err := s.db.GetNewsTopic(newsTopicID)
	if err != nil {
		slog.Error("Failed to get topic for source replacement", "topic_id", newsTopicID, "error", err)
		return
	}

	sourcingInstr, _ := s.db.GetSetting("news_sourcing_instructions")

	// Collect existing source URLs to avoid duplicates
	existingSources, _ := s.db.GetSourcesForNewsTopic(newsTopicID)
	existingURLs := make(map[string]bool, len(existingSources))
	for _, src := range existingSources {
		existingURLs[src.URL] = true
	}

	replaceCtx, replaceCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer replaceCancel()

	discovered, _, _, _, err := s.ai.DiscoverSources(replaceCtx, ai.DiscoverOpts{
		TopicName:            topic.Name,
		Description:          topic.Description,
		SourcingInstructions: sourcingInstr,
		AIProvider:           topic.AIProvider,
		IsNiche:              topic.IsNiche,
	})
	if err != nil {
		slog.Error("Failed to discover replacement sources", "topic", topic.Name, "error", err)
		return
	}

	added := 0
	for _, source := range discovered {
		if added >= count {
			break
		}
		if existingURLs[source.URL] {
			continue
		}
		if err := scraper.ValidateURL(source.URL); err != nil {
			continue
		}

		// Validate source: test-scrape + RSS auto-discovery
		result := s.scraper.ValidateSource(ctx, source.URL, source.Name)
		if !result.OK {
			slog.Info("Rejected replacement source (validation failed)",
				"url", source.URL, "name", source.Name, "reason", result.Reason)
			continue
		}

		finalURL := source.URL
		if result.FeedURL != "" {
			finalURL = result.FeedURL
		}

		if _, err := s.db.AddNewsSource(newsTopicID, finalURL, source.Name, false); err != nil {
			slog.Error("Failed to add replacement source", "error", err)
			continue
		}
		existingURLs[finalURL] = true
		added++
		slog.Info("Added replacement news source", "topic", topic.Name, "url", finalURL)
	}

	slog.Info("Replaced removed sources", "topic", topic.Name, "removed", count, "replaced", added)
}

// mineRedditDomains collects frequently-shared external domains from relevant subreddits.
// It checks existing topic sources and curated feeds for Reddit URLs, then mines their
// top link posts to find domains the community values.
func (s *Scheduler) mineRedditDomains(ctx context.Context, newsTopicID int64, topicName, description string) []string {
	// Collect Reddit URLs from existing sources and curated feeds
	var redditURLs []string

	existingSources, _ := s.db.GetSourcesForNewsTopic(newsTopicID)
	for _, src := range existingSources {
		if reddit.IsRedditURL(src.URL) {
			redditURLs = append(redditURLs, src.URL)
		}
	}

	for _, feed := range feeds.FindRelevant(topicName, description) {
		if reddit.IsRedditURL(feed.URL) {
			// Avoid duplicates
			found := false
			for _, u := range redditURLs {
				if u == feed.URL {
					found = true
					break
				}
			}
			if !found {
				redditURLs = append(redditURLs, feed.URL)
			}
		}
	}

	if len(redditURLs) == 0 {
		return nil
	}

	// Mine top link posts from up to 3 subreddits
	redditClient := reddit.New()
	var allLinkPosts []reddit.LinkPost
	limit := 3
	if len(redditURLs) < limit {
		limit = len(redditURLs)
	}
	for _, subURL := range redditURLs[:limit] {
		links, err := redditClient.FetchTopLinks(ctx, subURL, 25)
		if err != nil {
			slog.Debug("Failed to fetch Reddit links for mining", "url", subURL, "error", err)
			continue
		}
		allLinkPosts = append(allLinkPosts, links...)
	}

	if len(allLinkPosts) == 0 {
		return nil
	}

	ranked := reddit.RankDomains(allLinkPosts)

	// Take top 8 domains
	maxDomains := 8
	if len(ranked) < maxDomains {
		maxDomains = len(ranked)
	}

	var domains []string
	for _, r := range ranked[:maxDomains] {
		domains = append(domains, r.Domain)
	}

	slog.Info("Mined Reddit link domains", "topic", topicName, "domains", domains)
	return domains
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

// logNewsRefreshError logs a news refresh error to the refresh_log table.
func (s *Scheduler) logNewsRefreshError(topic models.NewsTopic, start time.Time, err error) {
	s.db.LogRefresh(models.RefreshLog{
		TopicType: "news", TopicID: topic.ID, TopicName: topic.Name,
		Status: "error", ErrorType: classifyError(err), ErrorMessage: err.Error(),
		DurationMs: time.Since(start).Milliseconds(),
		AIProvider: topic.AIProvider,
	})
}

// classifyError categorizes an error into a descriptive type for the refresh log.
func classifyError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()

	switch {
	case strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "context canceled"):
		return "timeout"
	case strings.Contains(msg, "model") && strings.Contains(msg, "not found"):
		return "model_not_found"
	case strings.Contains(msg, "no sources available"):
		return "no_sources"
	case strings.Contains(msg, "failed to scrape any content") || strings.Contains(msg, "insufficient content"):
		return "no_content"
	case strings.Contains(msg, "discover sources"):
		return "discovery_error"
	case strings.Contains(msg, "scrape error") || strings.Contains(msg, "failed to visit"):
		return "scrape_error"
	case strings.Contains(msg, "failed to parse") || strings.Contains(msg, "JSON"):
		return "parse_error"
	case strings.Contains(msg, "status 429") || strings.Contains(msg, "rate limit"):
		return "rate_limited"
	case strings.Contains(msg, "status 401") || strings.Contains(msg, "status 403") || strings.Contains(msg, "API key"):
		return "auth_error"
	case strings.Contains(msg, "status 400"):
		return "bad_request"
	case strings.Contains(msg, "status 404"):
		return "not_found"
	case strings.Contains(msg, "status 5"):
		return "server_error"
	case strings.Contains(msg, "empty response") || strings.Contains(msg, "no parseable facts"):
		return "empty_response"
	case strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host") || strings.Contains(msg, "dial tcp"):
		return "connection_error"
	case strings.Contains(msg, "summarize content"):
		return "summarize_error"
	case strings.Contains(msg, "panic"):
		return "panic"
	default:
		return "ai_error"
	}
}

// RefreshNewsNow triggers an immediate news topic refresh.
func (s *Scheduler) RefreshNewsNow(ctx context.Context, newsTopicID int64) {
	key := topicKey("news", newsTopicID)
	mu, ok := s.lockTopic(key)
	if !ok {
		slog.Warn("News topic is already being refreshed", "topic_id", newsTopicID)
		return
	}
	defer mu.Unlock()
	s.safeRefreshNewsTopic(ctx, newsTopicID)
}

// DiscoverSourcesNow triggers immediate source discovery for a news topic.
func (s *Scheduler) DiscoverSourcesNow(ctx context.Context, newsTopicID int64) error {
	key := topicKey("news", newsTopicID)
	mu, ok := s.lockTopic(key)
	if !ok {
		return fmt.Errorf("news topic is already being refreshed")
	}
	defer mu.Unlock()
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
