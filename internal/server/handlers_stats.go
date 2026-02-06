package server

import (
	"log/slog"
	"net/http"
)

func (s *Server) handleStatsPage(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.GetStats()
	if err != nil {
		slog.Error("Failed to get stats", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	recentUsage, err := s.db.RecentAPIUsage(20)
	if err != nil {
		slog.Error("Failed to get recent usage", "error", err)
	}

	data := map[string]any{
		"Page":        "stats",
		"Stats":       stats,
		"RecentUsage": recentUsage,
	}
	s.render(w, "stats", data)
}
