package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/thinkscotty/kibble/internal/models"
)

func (s *Server) handleFactCreate(w http.ResponseWriter, r *http.Request) {
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Fact content is required", 400)
		return
	}

	topicIDStr := r.FormValue("topic_id")
	topicID, err := strconv.ParseInt(topicIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	trigrams := s.sim.Trigrams(content)
	fact := &models.Fact{
		TopicID:  topicID,
		Content:  content,
		Trigrams: s.sim.TrigramsToJSON(trigrams),
		IsCustom: true,
		Source:   "user",
	}

	if err := s.db.CreateFact(fact); err != nil {
		slog.Error("Failed to create fact", "error", err)
		http.Error(w, "Failed to create fact", 500)
		return
	}

	s.renderPartial(w, "fact_item", fact)
}

func (s *Server) handleFactEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid fact ID", 400)
		return
	}

	fact, err := s.db.GetFact(id)
	if err != nil {
		http.Error(w, "Fact not found", 404)
		return
	}

	s.renderPartial(w, "fact_edit", &fact)
}

func (s *Server) handleFactUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid fact ID", 400)
		return
	}

	fact, err := s.db.GetFact(id)
	if err != nil {
		http.Error(w, "Fact not found", 404)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Fact content is required", 400)
		return
	}

	fact.Content = content
	trigrams := s.sim.Trigrams(content)
	fact.Trigrams = s.sim.TrigramsToJSON(trigrams)

	if err := s.db.UpdateFact(&fact); err != nil {
		slog.Error("Failed to update fact", "error", err)
		http.Error(w, "Failed to update fact", 500)
		return
	}

	s.renderPartial(w, "fact_item", &fact)
}

func (s *Server) handleFactDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid fact ID", 400)
		return
	}

	if err := s.db.DeleteFact(id); err != nil {
		slog.Error("Failed to delete fact", "error", err)
		http.Error(w, "Failed to delete fact", 500)
		return
	}

	// Return empty response for HTMX to remove the element
	w.WriteHeader(200)
}

func (s *Server) handleFactSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(200)
		return
	}

	var topicID *int64
	if v := r.URL.Query().Get("topic_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			topicID = &id
		}
	}

	facts, err := s.db.SearchFacts(query, topicID)
	if err != nil {
		slog.Error("Failed to search facts", "error", err)
		http.Error(w, "Search failed", 500)
		return
	}

	// Enrich with topic names
	topicNames := make(map[int64]string)
	topics, _ := s.db.ListTopics()
	for _, t := range topics {
		topicNames[t.ID] = t.Name
	}
	for i := range facts {
		facts[i].TopicName = topicNames[facts[i].TopicID]
	}

	s.renderPartial(w, "fact_search_results", facts)
}
