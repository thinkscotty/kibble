package server

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/thinkscotty/kibble/internal/auth"
	"github.com/thinkscotty/kibble/internal/models"
)

// isHTTPS checks if the original request was made over HTTPS by examining
// the X-Forwarded-Proto header (set by reverse proxies) or the TLS state.
func isHTTPS(r *http.Request) bool {
	// Check X-Forwarded-Proto header (set by Caddy/nginx/etc)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		return true
	}
	// Check if direct TLS connection
	return r.TLS != nil
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{"Page": "login"}
	s.render(w, "login", data)
}

func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	if username == "" || password == "" {
		s.render(w, "login", map[string]any{
			"Page":  "login",
			"Error": "Username and password are required",
		})
		return
	}

	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		slog.Debug("Login failed: user lookup", "username", username, "error", err)
		s.render(w, "login", map[string]any{
			"Page":  "login",
			"Error": "Invalid username or password",
		})
		return
	}

	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		slog.Debug("Login failed: wrong password", "username", username)
		s.render(w, "login", map[string]any{
			"Page":  "login",
			"Error": "Invalid username or password",
		})
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		slog.Error("Failed to generate session token", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	sess := &models.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.db.CreateSession(sess); err != nil {
		slog.Error("Failed to create session", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "kibble_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	slog.Info("User logged in", "username", username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("kibble_session"); err == nil {
		s.db.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "kibble_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleSetupPage(w http.ResponseWriter, r *http.Request) {
	count, _ := s.db.UserCount()
	if count > 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	s.render(w, "setup", map[string]any{"Page": "setup"})
}

func (s *Server) handleSetupSubmit(w http.ResponseWriter, r *http.Request) {
	count, _ := s.db.UserCount()
	if count > 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	if username == "" || password == "" {
		s.render(w, "setup", map[string]any{
			"Page":  "setup",
			"Error": "Username and password are required",
		})
		return
	}
	if len(password) < 8 {
		s.render(w, "setup", map[string]any{
			"Page":     "setup",
			"Error":    "Password must be at least 8 characters",
			"Username": username,
		})
		return
	}
	if password != confirm {
		s.render(w, "setup", map[string]any{
			"Page":     "setup",
			"Error":    "Passwords do not match",
			"Username": username,
		})
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	user := &models.User{
		Username:     username,
		PasswordHash: hash,
	}
	if err := s.db.CreateUser(user); err != nil {
		slog.Error("Failed to create user", "error", err)
		s.render(w, "setup", map[string]any{
			"Page":     "setup",
			"Error":    "Failed to create account",
			"Username": username,
		})
		return
	}

	s.hasUsers.Store(true)

	slog.Info("Admin account created", "username", username)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
