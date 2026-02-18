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

	refreshLogs, err := s.db.RecentRefreshLogs(50)
	if err != nil {
		slog.Error("Failed to get refresh logs", "error", err)
	}

	data := map[string]any{
		"Page":        "stats",
		"Stats":       stats,
		"RecentUsage": recentUsage,
		"RefreshLogs": refreshLogs,
	}
	s.render(w, "stats", data)
}
