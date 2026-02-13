package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/thinkscotty/kibble/internal/models"
	"github.com/thinkscotty/kibble/internal/scraper"
)

func (s *Server) handleNewsPage(w http.ResponseWriter, r *http.Request) {
	newsTopics, err := s.db.ListNewsTopics()
	if err != nil {
		slog.Error("Failed to list news topics", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	var topicsWithSources []models.NewsTopicWithSources
	for _, nt := range newsTopics {
		sources, _ := s.db.GetSourcesForNewsTopic(nt.ID)
		topicsWithSources = append(topicsWithSources, models.NewsTopicWithSources{
			NewsTopic: nt,
			Sources:   sources,
		})
	}

	settings, _ := s.db.GetAllSettings()

	data := map[string]any{
		"Page":       "news",
		"NewsTopics": topicsWithSources,
		"Settings":   settings,
	}
	s.render(w, "news", data)
}

func (s *Server) handleNewsTopicCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Topic name is required", 400)
		return
	}

	storiesPerRefresh := 5
	if v := r.FormValue("stories_per_refresh"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			storiesPerRefresh = n
		}
	}

	refreshInterval := 120
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

	nt := &models.NewsTopic{
		Name:                   name,
		Description:            r.FormValue("description"),
		IsActive:               true,
		StoriesPerRefresh:      storiesPerRefresh,
		RefreshIntervalMinutes: refreshInterval,
		SummaryMinWords:        summaryMinWords,
		SummaryMaxWords:        summaryMaxWords,
	}

	if err := s.db.CreateNewsTopic(nt); err != nil {
		slog.Error("Failed to create news topic", "error", err)
		http.Error(w, "Failed to create news topic", 500)
		return
	}

	// Trigger background source discovery
	go func() {
		if err := s.sched.DiscoverSourcesNow(r.Context(), nt.ID); err != nil {
			slog.Error("Background source discovery failed", "topic_id", nt.ID, "error", err)
		}
	}()

	data := models.NewsTopicWithSources{
		NewsTopic: *nt,
		Sources:   nil,
	}
	s.renderPartial(w, "news_topic_row", data)
}

func (s *Server) handleNewsTopicEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	nt, err := s.db.GetNewsTopic(id)
	if err != nil {
		http.Error(w, "News topic not found", 404)
		return
	}

	s.renderPartial(w, "news_topic_edit_row", &nt)
}

func (s *Server) handleNewsTopicUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	nt, err := s.db.GetNewsTopic(id)
	if err != nil {
		http.Error(w, "News topic not found", 404)
		return
	}

	if name := r.FormValue("name"); name != "" {
		nt.Name = name
	}
	nt.Description = r.FormValue("description")

	if v := r.FormValue("stories_per_refresh"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			nt.StoriesPerRefresh = n
		}
	}
	if v := r.FormValue("refresh_interval_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			nt.RefreshIntervalMinutes = n
		}
	}
	if v := r.FormValue("summary_min_words"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			nt.SummaryMinWords = n
		}
	}
	if v := r.FormValue("summary_max_words"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			nt.SummaryMaxWords = n
		}
	}

	if err := s.db.UpdateNewsTopic(&nt); err != nil {
		slog.Error("Failed to update news topic", "error", err)
		http.Error(w, "Failed to update news topic", 500)
		return
	}

	sources, _ := s.db.GetSourcesForNewsTopic(id)
	data := models.NewsTopicWithSources{
		NewsTopic: nt,
		Sources:   sources,
	}
	s.renderPartial(w, "news_topic_row", data)
}

func (s *Server) handleNewsTopicDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	if err := s.db.DeleteNewsTopic(id); err != nil {
		slog.Error("Failed to delete news topic", "error", err)
		http.Error(w, "Failed to delete news topic", 500)
		return
	}

	w.WriteHeader(200)
}

func (s *Server) handleNewsTopicToggle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	active := r.FormValue("active") == "true"

	if err := s.db.ToggleNewsTopicActive(id, active); err != nil {
		http.Error(w, "Failed to toggle news topic", 500)
		return
	}

	nt, _ := s.db.GetNewsTopic(id)
	sources, _ := s.db.GetSourcesForNewsTopic(id)
	data := models.NewsTopicWithSources{
		NewsTopic: nt,
		Sources:   sources,
	}
	s.renderPartial(w, "news_topic_row", data)
}

func (s *Server) handleNewsTopicRefresh(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	go s.sched.RefreshNewsNow(r.Context(), id)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<span class="text-success text-sm">Refresh started...</span>`)
}

func (s *Server) handleNewsTopicDiscover(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	if err := s.sched.DiscoverSourcesNow(r.Context(), id); err != nil {
		slog.Error("Source discovery failed", "error", err)
		http.Error(w, "Source discovery failed: "+err.Error(), 500)
		return
	}

	// Return updated sources list
	nt, _ := s.db.GetNewsTopic(id)
	sources, _ := s.db.GetSourcesForNewsTopic(id)
	data := models.NewsTopicWithSources{
		NewsTopic: nt,
		Sources:   sources,
	}
	s.renderPartial(w, "news_topic_row", data)
}

func (s *Server) handleNewsSourceAdd(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid topic ID", 400)
		return
	}

	url := r.FormValue("url")
	name := r.FormValue("name")
	if url == "" {
		http.Error(w, "URL is required", 400)
		return
	}
	if err := scraper.ValidateURL(url); err != nil {
		http.Error(w, "Invalid URL: "+err.Error(), 400)
		return
	}
	if name == "" {
		name = url
	}

	if _, err := s.db.AddNewsSource(id, url, name, true); err != nil {
		slog.Error("Failed to add news source", "error", err)
		http.Error(w, "Failed to add source", 500)
		return
	}

	// Return updated topic row with sources
	nt, _ := s.db.GetNewsTopic(id)
	sources, _ := s.db.GetSourcesForNewsTopic(id)
	data := models.NewsTopicWithSources{
		NewsTopic: nt,
		Sources:   sources,
	}
	s.renderPartial(w, "news_topic_row", data)
}

func (s *Server) handleNewsSourceDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid source ID", 400)
		return
	}

	if err := s.db.DeleteNewsSource(id); err != nil {
		slog.Error("Failed to delete news source", "error", err)
		http.Error(w, "Failed to delete source", 500)
		return
	}

	w.WriteHeader(200)
}
