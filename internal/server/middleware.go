package server

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		slog.Debug("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start).String(),
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "path", r.URL.Path)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// setupGuard redirects all requests to /setup when no users exist.
func (s *Server) setupGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		if s.hasUsers.Load() {
			next.ServeHTTP(w, r)
			return
		}

		count, _ := s.db.UserCount()
		if count > 0 {
			s.hasUsers.Store(true)
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/setup" {
			next.ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, "/setup", http.StatusSeeOther)
	})
}

// requireAuth checks for a valid session cookie and redirects to /login on failure.
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("kibble_session")
		if err != nil {
			s.authFailed(w, r)
			return
		}

		if _, err := s.db.GetSession(cookie.Value); err != nil {
			s.authFailed(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authFailed handles an unauthenticated request.
func (s *Server) authFailed(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// requireAPIKey checks for a valid API key via Bearer token or query parameter.
func (s *Server) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var providedKey string

		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			providedKey = strings.TrimPrefix(auth, "Bearer ")
		}
		if providedKey == "" {
			providedKey = r.URL.Query().Get("api_key")
		}

		if providedKey == "" {
			jsonError(w, "API key required", http.StatusUnauthorized)
			return
		}

		storedKey, err := s.db.GetSetting("api_key")
		if err != nil || storedKey == "" {
			slog.Error("API key not configured")
			jsonError(w, "API key not configured", http.StatusInternalServerError)
			return
		}

		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(storedKey)) != 1 {
			jsonError(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
