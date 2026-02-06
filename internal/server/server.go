package server

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	kibble "github.com/thinkscotty/kibble"
	"github.com/thinkscotty/kibble/internal/config"
	"github.com/thinkscotty/kibble/internal/database"
	"github.com/thinkscotty/kibble/internal/gemini"
	"github.com/thinkscotty/kibble/internal/scheduler"
	"github.com/thinkscotty/kibble/internal/similarity"
)

type Server struct {
	cfg      config.Config
	db       *database.DB
	gemini   *gemini.Client
	sim      *similarity.Checker
	sched    *scheduler.Scheduler
	pages    map[string]*template.Template
	partials *template.Template
	httpSrv  *http.Server
}

func New(cfg config.Config, db *database.DB, geminiClient *gemini.Client, sim *similarity.Checker, sched *scheduler.Scheduler) *Server {
	s := &Server{
		cfg:    cfg,
		db:     db,
		gemini: geminiClient,
		sim:    sim,
		sched:  sched,
	}
	return s
}

// Start loads templates, sets up routes, and starts the HTTP server.
func (s *Server) Start() error {
	if err := s.loadTemplates(); err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	mux := http.NewServeMux()
	s.routes(mux)

	handler := recoveryMiddleware(loggingMiddleware(mux))

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(s.cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(s.cfg.Server.WriteTimeoutSeconds) * time.Second,
	}

	slog.Info("Starting server", "addr", addr)
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) routes(mux *http.ServeMux) {
	// Static assets from embedded filesystem
	staticFS, _ := fs.Sub(kibble.StaticFS, "web/static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	// Pages (full HTML)
	mux.HandleFunc("GET /{$}", s.handleDashboard)
	mux.HandleFunc("GET /topics", s.handleTopicsPage)
	mux.HandleFunc("GET /settings", s.handleSettingsPage)
	mux.HandleFunc("GET /stats", s.handleStatsPage)

	// Topics HTMX endpoints (return partials)
	mux.HandleFunc("POST /topics", s.handleTopicCreate)
	mux.HandleFunc("GET /topics/{id}/edit", s.handleTopicEditForm)
	mux.HandleFunc("PUT /topics/{id}", s.handleTopicUpdate)
	mux.HandleFunc("DELETE /topics/{id}", s.handleTopicDelete)
	mux.HandleFunc("PATCH /topics/{id}/toggle", s.handleTopicToggle)
	mux.HandleFunc("POST /topics/reorder", s.handleTopicReorder)
	mux.HandleFunc("POST /topics/{id}/refresh", s.handleTopicRefresh)

	// Facts HTMX endpoints (return partials)
	mux.HandleFunc("POST /facts", s.handleFactCreate)
	mux.HandleFunc("GET /facts/{id}/edit", s.handleFactEditForm)
	mux.HandleFunc("PUT /facts/{id}", s.handleFactUpdate)
	mux.HandleFunc("DELETE /facts/{id}", s.handleFactDelete)
	mux.HandleFunc("GET /facts/search", s.handleFactSearch)

	// Settings HTMX endpoints
	mux.HandleFunc("POST /settings", s.handleSettingsUpdate)
	mux.HandleFunc("POST /settings/apikey/test", s.handleAPIKeyTest)

	// External Client API (JSON)
	mux.HandleFunc("GET /api/v1/topics", s.handleAPITopics)
	mux.HandleFunc("GET /api/v1/facts", s.handleAPIFacts)
	mux.HandleFunc("GET /api/v1/facts/random", s.handleAPIRandomFact)
}

func (s *Server) loadTemplates() error {
	funcMap := template.FuncMap{
		"safe": func(str string) template.HTML {
			return template.HTML(str)
		},
		"timeAgo": func(t *time.Time) string {
			if t == nil {
				return "Never"
			}
			d := time.Since(*t)
			switch {
			case d < time.Minute:
				return "Just now"
			case d < time.Hour:
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(d.Hours()))
			default:
				return fmt.Sprintf("%dd ago", int(d.Hours()/24))
			}
		},
		"boolChecked": func(b bool) string {
			if b {
				return "checked"
			}
			return ""
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"formatBytes": func(b int64) string {
			const unit = 1024
			if b < unit {
				return fmt.Sprintf("%d B", b)
			}
			div, exp := int64(unit), 0
			for n := b / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
		},
	}

	s.pages = make(map[string]*template.Template)

	pageNames := []string{"dashboard", "topics", "settings", "stats"}
	for _, page := range pageNames {
		t, err := template.New("base.html").Funcs(funcMap).ParseFS(kibble.TemplateFS,
			"web/templates/layouts/base.html",
			"web/templates/partials/*.html",
			"web/templates/pages/"+page+".html",
		)
		if err != nil {
			return fmt.Errorf("parse template %s: %w", page, err)
		}
		s.pages[page] = t
	}

	// Parse partials standalone for HTMX responses
	s.partials, _ = template.New("partials").Funcs(funcMap).ParseFS(kibble.TemplateFS,
		"web/templates/partials/*.html",
	)

	return nil
}

// render executes a full page template.
func (s *Server) render(w http.ResponseWriter, page string, data map[string]any) {
	tmpl, ok := s.pages[page]
	if !ok {
		http.Error(w, "Template not found", 500)
		return
	}

	// Inject settings into every page render
	if _, exists := data["Settings"]; !exists {
		settings, _ := s.db.GetAllSettings()
		data["Settings"] = settings
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		slog.Error("Template execution error", "page", page, "error", err)
	}
}

// renderPartial executes a named partial template for HTMX responses.
func (s *Server) renderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.partials.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("Partial execution error", "name", name, "error", err)
		http.Error(w, "Template error", 500)
	}
}
