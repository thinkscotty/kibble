package server

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
)

func (s *Server) handleAPITopics(w http.ResponseWriter, r *http.Request) {
	topics, err := s.db.ListActiveTopics()
	if err != nil {
		slog.Error("API: failed to list topics", "error", err)
		jsonError(w, "Failed to list topics", 500)
		return
	}

	type topicResp struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		FactCount int    `json:"fact_count"`
	}

	var result []topicResp
	for _, t := range topics {
		facts, _ := s.db.ListFactsByTopic(t.ID, 1000)
		result = append(result, topicResp{
			ID:        t.ID,
			Name:      t.Name,
			FactCount: len(facts),
		})
	}

	jsonResponse(w, map[string]any{"topics": result})
}

func (s *Server) handleAPIFacts(w http.ResponseWriter, r *http.Request) {
	topicIDStr := r.URL.Query().Get("topic_id")
	if topicIDStr == "" {
		jsonError(w, "topic_id parameter is required", 400)
		return
	}

	topicID, err := strconv.ParseInt(topicIDStr, 10, 64)
	if err != nil {
		jsonError(w, "Invalid topic_id", 400)
		return
	}

	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	topic, err := s.db.GetTopic(topicID)
	if err != nil {
		jsonError(w, "Topic not found", 404)
		return
	}

	facts, err := s.db.ListFactsByTopic(topicID, limit)
	if err != nil {
		slog.Error("API: failed to list facts", "error", err)
		jsonError(w, "Failed to list facts", 500)
		return
	}

	type factResp struct {
		ID      int64  `json:"id"`
		Content string `json:"content"`
	}

	var factList []factResp
	for _, f := range facts {
		factList = append(factList, factResp{ID: f.ID, Content: f.Content})
	}

	jsonResponse(w, map[string]any{
		"topic": topic.Name,
		"facts": factList,
	})
}

func (s *Server) handleAPIRandomFact(w http.ResponseWriter, r *http.Request) {
	topics, err := s.db.ListActiveTopics()
	if err != nil || len(topics) == 0 {
		jsonError(w, "No active topics found", 404)
		return
	}

	// Collect all facts from active topics
	type factWithTopic struct {
		ID      int64  `json:"id"`
		Topic   string `json:"topic"`
		Content string `json:"content"`
	}

	var allFacts []factWithTopic
	for _, t := range topics {
		facts, _ := s.db.ListFactsByTopic(t.ID, 100)
		for _, f := range facts {
			allFacts = append(allFacts, factWithTopic{
				ID:      f.ID,
				Topic:   t.Name,
				Content: f.Content,
			})
		}
	}

	if len(allFacts) == 0 {
		jsonError(w, "No facts available", 404)
		return
	}

	chosen := allFacts[rand.Intn(len(allFacts))]
	jsonResponse(w, map[string]any{"fact": chosen})
}

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
