package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
	path string
}

func New(path string) (*DB, error) {
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
	}

	for _, stmt := range statements {
		if _, err := db.conn.Exec(stmt); err != nil {
			return fmt.Errorf("exec migration: %w\nstatement: %s", err, stmt)
		}
	}

	return db.seedSettings()
}

func (db *DB) seedSettings() error {
	defaults := map[string]string{
		"gemini_api_key":          "",
		"ai_custom_instructions":  "",
		"ai_tone_instructions":    "",
		"theme_mode":              "dark",
		"primary_color":           "#4F46E5",
		"secondary_color":         "#10B981",
		"text_size":               "medium",
		"card_columns":            "3",
		"facts_per_topic_display": "5",
		"similarity_threshold":    "0.6",
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
	return nil
}
