package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/thinkscotty/kibble/internal/models"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	topics, err := s.db.ListActiveTopics()
	if err != nil {
		slog.Error("Failed to list active topics", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	settings, _ := s.db.GetAllSettings()
	limit := 5
	if v, ok := settings["facts_per_topic_display"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	var topicsWithFacts []models.TopicWithFacts
	for _, topic := range topics {
		facts, err := s.db.ListFactsByTopic(topic.ID, limit)
		if err != nil {
			slog.Error("Failed to list facts", "topic_id", topic.ID, "error", err)
			continue
		}
		topicsWithFacts = append(topicsWithFacts, models.TopicWithFacts{
			Topic: topic,
			Facts: facts,
		})
	}

	// Load news topics with stories
	storiesLimit := 5
	if v, ok := settings["stories_per_topic_display"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			storiesLimit = n
		}
	}

	activeNewsTopics, _ := s.db.ListActiveNewsTopics()
	var newsTopicsWithStories []models.NewsTopicWithStories
	for _, nt := range activeNewsTopics {
		stories, err := s.db.ListStoriesByNewsTopic(nt.ID, storiesLimit)
		if err != nil {
			slog.Error("Failed to list stories", "topic_id", nt.ID, "error", err)
			continue
		}
		newsTopicsWithStories = append(newsTopicsWithStories, models.NewsTopicWithStories{
			NewsTopic: nt,
			Stories:   stories,
		})
	}

	data := map[string]any{
		"Page":       "dashboard",
		"Topics":     topicsWithFacts,
		"NewsTopics": newsTopicsWithStories,
		"Settings":   settings,
	}

	s.render(w, "dashboard", data)
}
