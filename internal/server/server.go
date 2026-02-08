package server

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	kibble "github.com/thinkscotty/kibble"
	"github.com/thinkscotty/kibble/internal/config"
	"github.com/thinkscotty/kibble/internal/database"
	"github.com/thinkscotty/kibble/internal/gemini"
	"github.com/thinkscotty/kibble/internal/scheduler"
	"github.com/thinkscotty/kibble/internal/similarity"
)

type Server struct {
	cfg       config.Config
	db        *database.DB
	gemini    *gemini.Client
	sim       *similarity.Checker
	sched     *scheduler.Scheduler
	themes    []config.Theme
	hasUsers  atomic.Bool
	version   string
	buildTime string
	pages     map[string]*template.Template
	partials  *template.Template
	httpSrv   *http.Server
}

func New(cfg config.Config, db *database.DB, geminiClient *gemini.Client, sim *similarity.Checker, sched *scheduler.Scheduler, themes []config.Theme, version, buildTime string) *Server {
	s := &Server{
		cfg:       cfg,
		db:        db,
		gemini:    geminiClient,
		sim:       sim,
		sched:     sched,
		themes:    themes,
		version:   version,
		buildTime: buildTime,
	}
	if count, _ := db.UserCount(); count > 0 {
		s.hasUsers.Store(true)
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

	handler := recoveryMiddleware(loggingMiddleware(s.setupGuard(mux)))

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
	// Static assets — always public
	staticFS, _ := fs.Sub(kibble.StaticFS, "web/static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	// Auth routes — public
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("POST /login", s.handleLoginSubmit)
	mux.HandleFunc("POST /logout", s.handleLogout)
	mux.HandleFunc("GET /setup", s.handleSetupPage)
	mux.HandleFunc("POST /setup", s.handleSetupSubmit)

	// External Client API — protected by API key
	mux.Handle("GET /api/v1/topics", s.requireAPIKey(http.HandlerFunc(s.handleAPITopics)))
	mux.Handle("GET /api/v1/facts", s.requireAPIKey(http.HandlerFunc(s.handleAPIFacts)))
	mux.Handle("GET /api/v1/facts/random", s.requireAPIKey(http.HandlerFunc(s.handleAPIRandomFact)))

	// All other routes — protected by session auth
	mux.Handle("GET /{$}", s.requireAuth(http.HandlerFunc(s.handleDashboard)))
	mux.Handle("GET /topics", s.requireAuth(http.HandlerFunc(s.handleTopicsPage)))
	mux.Handle("GET /settings", s.requireAuth(http.HandlerFunc(s.handleSettingsPage)))
	mux.Handle("GET /stats", s.requireAuth(http.HandlerFunc(s.handleStatsPage)))

	mux.Handle("POST /topics", s.requireAuth(http.HandlerFunc(s.handleTopicCreate)))
	mux.Handle("GET /topics/{id}/edit", s.requireAuth(http.HandlerFunc(s.handleTopicEditForm)))
	mux.Handle("PUT /topics/{id}", s.requireAuth(http.HandlerFunc(s.handleTopicUpdate)))
	mux.Handle("DELETE /topics/{id}", s.requireAuth(http.HandlerFunc(s.handleTopicDelete)))
	mux.Handle("PATCH /topics/{id}/toggle", s.requireAuth(http.HandlerFunc(s.handleTopicToggle)))
	mux.Handle("POST /topics/reorder", s.requireAuth(http.HandlerFunc(s.handleTopicReorder)))
	mux.Handle("POST /topics/{id}/refresh", s.requireAuth(http.HandlerFunc(s.handleTopicRefresh)))

	mux.Handle("POST /facts", s.requireAuth(http.HandlerFunc(s.handleFactCreate)))
	mux.Handle("GET /facts/{id}/edit", s.requireAuth(http.HandlerFunc(s.handleFactEditForm)))
	mux.Handle("PUT /facts/{id}", s.requireAuth(http.HandlerFunc(s.handleFactUpdate)))
	mux.Handle("DELETE /facts/{id}", s.requireAuth(http.HandlerFunc(s.handleFactDelete)))
	mux.Handle("GET /facts/search", s.requireAuth(http.HandlerFunc(s.handleFactSearch)))

	mux.Handle("POST /settings", s.requireAuth(http.HandlerFunc(s.handleSettingsUpdate)))
	mux.Handle("POST /settings/apikey/test", s.requireAuth(http.HandlerFunc(s.handleAPIKeyTest)))
	mux.Handle("POST /settings/apikey/regenerate", s.requireAuth(http.HandlerFunc(s.handleAPIKeyRegenerate)))
	mux.Handle("POST /settings/update/check", s.requireAuth(http.HandlerFunc(s.handleUpdateCheck)))
	mux.Handle("POST /settings/update/install", s.requireAuth(http.HandlerFunc(s.handleUpdateInstall)))
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

	pageNames := []string{"dashboard", "topics", "settings", "stats", "login", "setup"}
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

	// Inject version info
	data["Version"] = s.version
	data["BuildTime"] = s.buildTime

	// Resolve the active theme and inject CSS variables + logo choice
	s.injectThemeData(data)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		slog.Error("Template execution error", "page", page, "error", err)
	}
}

// injectThemeData resolves the selected theme and adds ThemeCSS and ThemeLogo to the data map.
func (s *Server) injectThemeData(data map[string]any) {
	settings, _ := data["Settings"].(map[string]string)
	themeID := ""
	if settings != nil {
		themeID = settings["theme_mode"]
	}

	theme := s.findTheme(themeID)
	data["ThemeCSS"] = template.CSS(config.ResolveThemeCSS(theme))
	data["ThemeLogo"] = theme.Logo
	data["Themes"] = s.themes
	data["CurrentTheme"] = theme.ID
}

// findTheme looks up a theme by ID, falling back to the first available theme.
func (s *Server) findTheme(id string) config.Theme {
	for _, t := range s.themes {
		if t.ID == id {
			return t
		}
	}
	if len(s.themes) > 0 {
		return s.themes[0]
	}
	// Absolute fallback if no themes configured at all
	return config.DefaultThemes()[0]
}

// renderPartial executes a named partial template for HTMX responses.
func (s *Server) renderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.partials.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("Partial execution error", "name", name, "error", err)
		http.Error(w, "Template error", 500)
	}
}
