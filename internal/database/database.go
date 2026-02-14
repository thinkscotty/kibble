package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/thinkscotty/kibble/internal/apikey"
	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
	path string
}

func New(path string) (*DB, error) {
	// If path points to an existing directory, append the default filename.
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, "kibble.db")
		slog.Info("Database path is a directory, using file inside it", "path", path)
	}

	// Ensure the parent directory exists so SQLite can create the file.
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create database directory %q: %w", dir, err)
		}
	}

	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	conn.SetMaxOpenConns(2)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db := &DB{conn: conn, path: path}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// DatabaseSizeBytes returns the file size of the database.
func (db *DB) DatabaseSizeBytes() (int64, error) {
	info, err := os.Stat(db.path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func parseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

func (db *DB) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS topics (
			id                       INTEGER PRIMARY KEY AUTOINCREMENT,
			name                     TEXT    NOT NULL,
			description              TEXT    NOT NULL DEFAULT '',
			display_order            INTEGER NOT NULL DEFAULT 0,
			is_active                INTEGER NOT NULL DEFAULT 1,
			facts_per_refresh        INTEGER NOT NULL DEFAULT 5,
			refresh_interval_minutes INTEGER NOT NULL DEFAULT 1440,
			last_refreshed_at        TEXT,
			created_at               TEXT    NOT NULL DEFAULT (datetime('now')),
			updated_at               TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS facts (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			topic_id    INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
			content     TEXT    NOT NULL,
			trigrams    TEXT    NOT NULL DEFAULT '',
			is_custom   INTEGER NOT NULL DEFAULT 0,
			is_archived INTEGER NOT NULL DEFAULT 0,
			source      TEXT    NOT NULL DEFAULT 'gemini',
			created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
			updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_facts_topic_id ON facts(topic_id)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key        TEXT PRIMARY KEY,
			value      TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS api_usage_log (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			topic_id        INTEGER REFERENCES topics(id) ON DELETE SET NULL,
			facts_requested INTEGER NOT NULL DEFAULT 0,
			facts_generated INTEGER NOT NULL DEFAULT 0,
			facts_discarded INTEGER NOT NULL DEFAULT 0,
			tokens_used     INTEGER NOT NULL DEFAULT 0,
			error_message   TEXT,
			created_at      TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			username      TEXT    NOT NULL UNIQUE,
			password_hash TEXT    NOT NULL,
			created_at    TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			token      TEXT    NOT NULL UNIQUE,
			user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at TEXT    NOT NULL,
			created_at TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)`,

		// News / Updates feature
		`CREATE TABLE IF NOT EXISTS news_topics (
			id                       INTEGER PRIMARY KEY AUTOINCREMENT,
			name                     TEXT    NOT NULL,
			description              TEXT    NOT NULL DEFAULT '',
			display_order            INTEGER NOT NULL DEFAULT 0,
			is_active                INTEGER NOT NULL DEFAULT 1,
			stories_per_refresh      INTEGER NOT NULL DEFAULT 5,
			refresh_interval_minutes INTEGER NOT NULL DEFAULT 120,
			last_refreshed_at        TEXT,
			created_at               TEXT    NOT NULL DEFAULT (datetime('now')),
			updated_at               TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS news_sources (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			news_topic_id  INTEGER NOT NULL REFERENCES news_topics(id) ON DELETE CASCADE,
			url            TEXT    NOT NULL,
			name           TEXT    NOT NULL,
			is_manual      INTEGER NOT NULL DEFAULT 0,
			is_active      INTEGER NOT NULL DEFAULT 1,
			failure_count  INTEGER NOT NULL DEFAULT 0,
			last_error     TEXT    NOT NULL DEFAULT '',
			created_at     TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_news_sources_topic ON news_sources(news_topic_id)`,
		`CREATE TABLE IF NOT EXISTS stories (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			news_topic_id  INTEGER NOT NULL REFERENCES news_topics(id) ON DELETE CASCADE,
			title          TEXT    NOT NULL,
			summary        TEXT    NOT NULL,
			source_url     TEXT    NOT NULL DEFAULT '',
			source_title   TEXT    NOT NULL DEFAULT '',
			published_at   TEXT    NOT NULL DEFAULT (datetime('now')),
			created_at     TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_stories_topic ON stories(news_topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stories_created ON stories(created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS news_refresh_status (
			news_topic_id  INTEGER PRIMARY KEY REFERENCES news_topics(id) ON DELETE CASCADE,
			last_refresh   TEXT,
			next_refresh   TEXT,
			status         TEXT    NOT NULL DEFAULT 'pending',
			error_message  TEXT    NOT NULL DEFAULT ''
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.conn.Exec(stmt); err != nil {
			return fmt.Errorf("exec migration: %w\nstatement: %s", err, stmt)
		}
	}

	// Additive migrations (safe to re-run; ALTER TABLE errors ignored for existing columns)
	alterStatements := []string{
		`ALTER TABLE topics ADD COLUMN summary_min_words INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE topics ADD COLUMN summary_max_words INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE news_topics ADD COLUMN summary_min_words INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE news_topics ADD COLUMN summary_max_words INTEGER NOT NULL DEFAULT 0`,
	}
	for _, stmt := range alterStatements {
		db.conn.Exec(stmt) // ignore "duplicate column" errors
	}

	return db.seedSettings()
}

func (db *DB) seedSettings() error {
	defaults := map[string]string{
		"gemini_api_key":          "",
		"ai_custom_instructions":  "",
		"ai_tone_instructions":    "",
		"theme_mode":              "soft-dark",
		"text_size":               "medium",
		"card_columns":            "3",
		"facts_per_topic_display": "5",
		"similarity_threshold":    "0.6",
		"news_sourcing_instructions":    "Find reliable, reputable news sources that provide regular updates. Include relevant Reddit subreddits when appropriate. Prefer sources with RSS feeds or well-structured HTML. Avoid paywalled content when possible.",
		"news_summarizing_instructions": "Summarize the news story in a clear, informative tone. Focus on the key facts and why this story matters. Keep the summary between 75-150 words.",
		"news_tone_instructions":        "",
		"stories_per_topic_display":     "5",
	}

	stmt, err := db.conn.Prepare(`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range defaults {
		if _, err := stmt.Exec(key, value); err != nil {
			return err
		}
	}

	// Seed a random API key if one doesn't already exist.
	var apiKeyExists int
	db.conn.QueryRow(`SELECT COUNT(*) FROM settings WHERE key = 'api_key'`).Scan(&apiKeyExists)
	if apiKeyExists == 0 {
		key, err := apikey.Generate()
		if err != nil {
			return fmt.Errorf("generate api key: %w", err)
		}
		if _, err := db.conn.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('api_key', ?)`, key); err != nil {
			return err
		}
	}

	return nil
}
