package server

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/thinkscotty/kibble/internal/apikey"
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

	// Check if the currently selected theme exists
	if themeID := settings["theme_mode"]; themeID != "" {
		found := false
		for _, t := range s.themes {
			if t.ID == themeID {
				found = true
				break
			}
		}
		if !found {
			// Theme was auto-corrected by findTheme - notify user
			data["Warning"] = fmt.Sprintf("Your previous theme was removed in an update. Switched to the default theme.")
		}
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
		"ai_provider",
		"ollama_url",
		"ollama_model",
		"ai_custom_instructions",
		"ai_tone_instructions",
		"news_sourcing_instructions",
		"news_summarizing_instructions",
		"news_tone_instructions",
		"theme_mode",
		"text_size",
		"card_columns",
		"facts_per_topic_display",
		"stories_per_topic_display",
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

	err := s.ai.TestGeminiKey(r.Context(), apiKey)
	if err != nil {
		slog.Error("API key test failed", "error", err)
		w.Write([]byte(`<span class="text-error">API key test failed: ` + err.Error() + `</span>`))
		return
	}

	w.Write([]byte(`<span class="text-success">API key is valid!</span>`))
}

func (s *Server) handleOllamaTest(w http.ResponseWriter, r *http.Request) {
	ollamaURL := r.FormValue("ollama_url")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	// Temporarily save the URL so TestOllamaConnection can read it
	s.db.SetSetting("ollama_url", ollamaURL)

	err := s.ai.TestOllamaConnection(r.Context())
	if err != nil {
		slog.Error("Ollama connection test failed", "error", err)
		w.Write([]byte(`<span class="text-error">Connection failed: ` + template.HTMLEscapeString(err.Error()) + `</span>`))
		return
	}

	w.Write([]byte(`<span class="text-success">Connected to Ollama!</span>`))
}

func (s *Server) handleOllamaModels(w http.ResponseWriter, r *http.Request) {
	ollamaURL := r.FormValue("ollama_url")
	if ollamaURL != "" {
		s.db.SetSetting("ollama_url", ollamaURL)
	}

	models, err := s.ai.ListOllamaModels(r.Context())
	if err != nil {
		slog.Error("Failed to list Ollama models", "error", err)
		w.Write([]byte(`<option value="">Failed to load models</option>`))
		return
	}

	if len(models) == 0 {
		w.Write([]byte(`<option value="">No models found</option>`))
		return
	}

	currentModel, _ := s.db.GetSetting("ollama_model")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, m := range models {
		selected := ""
		if m.Name == currentModel {
			selected = " selected"
		}
		fmt.Fprintf(w, `<option value="%s"%s>%s</option>`, template.HTMLEscapeString(m.Name), selected, template.HTMLEscapeString(m.Name))
	}
}

func (s *Server) handleAPIKeyRegenerate(w http.ResponseWriter, r *http.Request) {
	newKey, err := apikey.Generate()
	if err != nil {
		slog.Error("Failed to generate API key", "error", err)
		http.Error(w, "Failed to generate key", 500)
		return
	}

	if err := s.db.SetSetting("api_key", newKey); err != nil {
		slog.Error("Failed to save API key", "error", err)
		http.Error(w, "Failed to save key", 500)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<code id="api-key-value" style="word-break: break-all;">%s</code>
		<span class="text-success text-sm" style="margin-left: 0.5rem;">Key regenerated!</span>`,
		template.HTMLEscapeString(newKey))
}
