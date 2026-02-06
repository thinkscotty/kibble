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

	data := map[string]any{
		"Page":     "dashboard",
		"Topics":   topicsWithFacts,
		"Settings": settings,
	}

	s.render(w, "dashboard", data)
}
