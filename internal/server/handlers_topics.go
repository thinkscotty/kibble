package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/thinkscotty/kibble/internal/models"
)

func (s *Server) handleTopicsPage(w http.ResponseWriter, r *http.Request) {
	topics, err := s.db.ListTopics()
	if err != nil {
		slog.Error("Failed to list topics", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	data := map[string]any{
		"Page":   "topics",
		"Topics": topics,
	}
	s.render(w, "topics", data)
}

func (s *Server) handleTopicCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Topic name is required", 400)
		return
	}

	factsPerRefresh := 5
	if v := r.FormValue("facts_per_refresh"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			factsPerRefresh = n
		}
	}

	refreshInterval := 1440
	if v := r.FormValue("refresh_interval_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			refreshInterval = n
		}
	}

	var summaryMinWords, summaryMaxWords int
	if v := r.FormValue("summary_min_words"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			summaryMinWords = n
		}
	}
	if v := r.FormValue("summary_max_words"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			summaryMaxWords = n
		}
	}

	topic := &models.Topic{
		Name:                   name,
		Description:            r.FormValue("description"),
		IsActive:               true,
		FactsPerRefresh:        factsPerRefresh,
		RefreshIntervalMinutes: refreshInterval,
		SummaryMinWords:        summaryMinWords,
		SummaryMaxWords:        summaryMaxWords,
		AIProvider:             r.FormValue("ai_provider"),
		IsNiche:                r.FormValue("is_niche") == "1",
	}

	if err := s.db.CreateTopic(topic); err != nil {
		slog.Error("Failed to create topic", "error", err)
		http.Error(w, "Failed to create topic", 500)
		return
	}

	s.renderPartial(w, "topic_row", topic)
}

func (s *Server) handleTopicEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	topic, err := s.db.GetTopic(id)
	if err != nil {
		http.Error(w, "Topic not found", 404)
		return
	}

	s.renderPartial(w, "topic_edit_row", &topic)
}

func (s *Server) handleTopicUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	topic, err := s.db.GetTopic(id)
	if err != nil {
		http.Error(w, "Topic not found", 404)
		return
	}

	if name := r.FormValue("name"); name != "" {
		topic.Name = name
	}
	topic.Description = r.FormValue("description")

	if v := r.FormValue("facts_per_refresh"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topic.FactsPerRefresh = n
		}
	}
	if v := r.FormValue("refresh_interval_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topic.RefreshIntervalMinutes = n
		}
	}
	if v := r.FormValue("summary_min_words"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			topic.SummaryMinWords = n
		}
	}
	if v := r.FormValue("summary_max_words"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			topic.SummaryMaxWords = n
		}
	}
	topic.AIProvider = r.FormValue("ai_provider")
	topic.IsNiche = r.FormValue("is_niche") == "1"

	if err := s.db.UpdateTopic(&topic); err != nil {
		slog.Error("Failed to update topic", "error", err)
		http.Error(w, "Failed to update topic", 500)
		return
	}

	s.renderPartial(w, "topic_row", &topic)
}

func (s *Server) handleTopicDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	if err := s.db.DeleteTopic(id); err != nil {
		slog.Error("Failed to delete topic", "error", err)
		http.Error(w, "Failed to delete topic", 500)
		return
	}

	w.WriteHeader(200)
}

func (s *Server) handleTopicToggle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	active := r.FormValue("active") == "true"

	if err := s.db.ToggleTopicActive(id, active); err != nil {
		http.Error(w, "Failed to toggle topic", 500)
		return
	}

	topic, _ := s.db.GetTopic(id)
	s.renderPartial(w, "topic_row", &topic)
}

func (s *Server) handleTopicReorder(w http.ResponseWriter, r *http.Request) {
	var ids []int64
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}

	if err := s.db.ReorderTopics(ids); err != nil {
		slog.Error("Failed to reorder topics", "error", err)
		http.Error(w, "Failed to reorder topics", 500)
		return
	}

	w.WriteHeader(200)
}

func (s *Server) handleTopicRefresh(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	if err := s.sched.RefreshNow(r.Context(), id); err != nil {
		slog.Error("Failed to refresh topic", "error", err)
		http.Error(w, "Failed to refresh: "+err.Error(), 500)
		return
	}

	// Return updated topic card for the dashboard
	topic, _ := s.db.GetTopic(id)
	settings, _ := s.db.GetAllSettings()
	limit := 5
	if v, ok := settings["facts_per_topic_display"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	facts, _ := s.db.ListFactsByTopic(id, limit)
	data := models.TopicWithFacts{Topic: topic, Facts: facts}
	s.renderPartial(w, "topic_card", data)
}
