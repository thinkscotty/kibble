package server

import (
	"log/slog"
	"net/http"
)

func (s *Server) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	settings, err := s.db.GetAllSettings()
	if err != nil {
		slog.Error("Failed to load settings", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	data := map[string]any{
		"Page":     "settings",
		"Settings": settings,
	}
	s.render(w, "settings", data)
}

func (s *Server) handleSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", 400)
		return
	}

	settingsKeys := []string{
		"gemini_api_key",
		"ai_custom_instructions",
		"ai_tone_instructions",
		"theme_mode",
		"primary_color",
		"secondary_color",
		"text_size",
		"card_columns",
		"facts_per_topic_display",
		"similarity_threshold",
	}

	for _, key := range settingsKeys {
		if value := r.FormValue(key); value != "" {
			if err := s.db.SetSetting(key, value); err != nil {
				slog.Error("Failed to save setting", "key", key, "error", err)
			}
		}
	}

	// For theme_mode, also handle unchecked case (form won't send value)
	if r.Form.Has("theme_mode") {
		s.db.SetSetting("theme_mode", r.FormValue("theme_mode"))
	}

	// Return success indicator for HTMX
	w.Header().Set("HX-Trigger", "settings-saved")
	settings, _ := s.db.GetAllSettings()
	data := map[string]any{
		"Page":     "settings",
		"Settings": settings,
		"Success":  "Settings saved successfully",
	}
	s.render(w, "settings", data)
}

func (s *Server) handleAPIKeyTest(w http.ResponseWriter, r *http.Request) {
	apiKey := r.FormValue("gemini_api_key")
	if apiKey == "" {
		w.Write([]byte(`<span class="text-error">Please enter an API key first</span>`))
		return
	}

	err := s.gemini.TestAPIKey(r.Context(), apiKey)
	if err != nil {
		slog.Error("API key test failed", "error", err)
		w.Write([]byte(`<span class="text-error">API key test failed: ` + err.Error() + `</span>`))
		return
	}

	w.Write([]byte(`<span class="text-success">API key is valid!</span>`))
}
