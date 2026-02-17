package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/thinkscotty/kibble/internal/models"
)

func (db *DB) ListTopics() ([]models.Topic, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, display_order, is_active, facts_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM topics ORDER BY display_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTopics(rows)
}

func (db *DB) ListActiveTopics() ([]models.Topic, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, display_order, is_active, facts_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM topics WHERE is_active = 1 ORDER BY display_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTopics(rows)
}

func (db *DB) GetTopic(id int64) (models.Topic, error) {
	var t models.Topic
	var lastRefreshed sql.NullString
	var createdAt, updatedAt string

	err := db.conn.QueryRow(`
		SELECT id, name, description, display_order, is_active, facts_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM topics WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Description, &t.DisplayOrder, &t.IsActive,
		&t.FactsPerRefresh, &t.RefreshIntervalMinutes,
		&t.SummaryMinWords, &t.SummaryMaxWords,
		&t.AIProvider, &t.IsNiche, &lastRefreshed,
		&createdAt, &updatedAt)
	if err != nil {
		return t, err
	}

	t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if lastRefreshed.Valid {
		parsed, _ := time.Parse("2006-01-02 15:04:05", lastRefreshed.String)
		t.LastRefreshedAt = &parsed
	}
	return t, nil
}

func (db *DB) CreateTopic(t *models.Topic) error {
	// Get next display_order
	var maxOrder sql.NullInt64
	db.conn.QueryRow(`SELECT MAX(display_order) FROM topics`).Scan(&maxOrder)
	nextOrder := 0
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := db.conn.Exec(`
		INSERT INTO topics (name, description, display_order, is_active, facts_per_refresh, refresh_interval_minutes, summary_min_words, summary_max_words, ai_provider, is_niche)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Name, t.Description, nextOrder, boolToInt(t.IsActive),
		t.FactsPerRefresh, t.RefreshIntervalMinutes,
		t.SummaryMinWords, t.SummaryMaxWords,
		t.AIProvider, boolToInt(t.IsNiche))
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id
	t.DisplayOrder = nextOrder
	return nil
}

func (db *DB) UpdateTopic(t *models.Topic) error {
	_, err := db.conn.Exec(`
		UPDATE topics SET name = ?, description = ?, is_active = ?,
		       facts_per_refresh = ?, refresh_interval_minutes = ?,
		       summary_min_words = ?, summary_max_words = ?,
		       ai_provider = ?, is_niche = ?,
		       updated_at = datetime('now')
		WHERE id = ?`,
		t.Name, t.Description, boolToInt(t.IsActive),
		t.FactsPerRefresh, t.RefreshIntervalMinutes,
		t.SummaryMinWords, t.SummaryMaxWords,
		t.AIProvider, boolToInt(t.IsNiche), t.ID)
	return err
}

func (db *DB) DeleteTopic(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM topics WHERE id = ?`, id)
	return err
}

func (db *DB) ToggleTopicActive(id int64, active bool) error {
	_, err := db.conn.Exec(`UPDATE topics SET is_active = ?, updated_at = datetime('now') WHERE id = ?`,
		boolToInt(active), id)
	return err
}

func (db *DB) ReorderTopics(ids []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE topics SET display_order = ?, updated_at = datetime('now') WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, id := range ids {
		if _, err := stmt.Exec(i, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) UpdateTopicRefreshTime(id int64) error {
	_, err := db.conn.Exec(`UPDATE topics SET last_refreshed_at = datetime('now'), updated_at = datetime('now') WHERE id = ?`, id)
	return err
}

func (db *DB) TopicsDueForRefresh() ([]models.Topic, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, display_order, is_active, facts_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM topics
		WHERE is_active = 1
		  AND (last_refreshed_at IS NULL
		       OR datetime('now') > datetime(last_refreshed_at, '+' || refresh_interval_minutes || ' minutes'))
		ORDER BY last_refreshed_at ASC NULLS FIRST`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTopics(rows)
}

func (db *DB) TopicCount() (total int, active int, err error) {
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM topics`).Scan(&total)
	if err != nil {
		return
	}
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM topics WHERE is_active = 1`).Scan(&active)
	return
}

func scanTopics(rows *sql.Rows) ([]models.Topic, error) {
	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		var lastRefreshed sql.NullString
		var createdAt, updatedAt string

		if err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.DisplayOrder, &t.IsActive,
			&t.FactsPerRefresh, &t.RefreshIntervalMinutes,
			&t.SummaryMinWords, &t.SummaryMaxWords,
			&t.AIProvider, &t.IsNiche, &lastRefreshed,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan topic: %w", err)
		}

		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		if lastRefreshed.Valid {
			parsed, _ := time.Parse("2006-01-02 15:04:05", lastRefreshed.String)
			t.LastRefreshedAt = &parsed
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
