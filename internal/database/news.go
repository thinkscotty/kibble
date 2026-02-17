package database

import (
	"database/sql"
	"fmt"

	"github.com/thinkscotty/kibble/internal/models"
)

// --- News Topics ---

func (db *DB) ListNewsTopics() ([]models.NewsTopic, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, display_order, is_active, stories_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM news_topics ORDER BY display_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNewsTopics(rows)
}

func (db *DB) ListActiveNewsTopics() ([]models.NewsTopic, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, display_order, is_active, stories_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM news_topics WHERE is_active = 1 ORDER BY display_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNewsTopics(rows)
}

func (db *DB) GetNewsTopic(id int64) (models.NewsTopic, error) {
	var t models.NewsTopic
	var lastRefreshed sql.NullString
	var createdAt, updatedAt string

	err := db.conn.QueryRow(`
		SELECT id, name, description, display_order, is_active, stories_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM news_topics WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Description, &t.DisplayOrder, &t.IsActive,
		&t.StoriesPerRefresh, &t.RefreshIntervalMinutes,
		&t.SummaryMinWords, &t.SummaryMaxWords,
		&t.AIProvider, &t.IsNiche, &lastRefreshed,
		&createdAt, &updatedAt)
	if err != nil {
		return t, err
	}

	t.CreatedAt, _ = parseTime(createdAt)
	t.UpdatedAt, _ = parseTime(updatedAt)
	if lastRefreshed.Valid {
		parsed, _ := parseTime(lastRefreshed.String)
		t.LastRefreshedAt = &parsed
	}
	return t, nil
}

func (db *DB) CreateNewsTopic(t *models.NewsTopic) error {
	var maxOrder sql.NullInt64
	db.conn.QueryRow(`SELECT MAX(display_order) FROM news_topics`).Scan(&maxOrder)
	nextOrder := 0
	if maxOrder.Valid {
		nextOrder = int(maxOrder.Int64) + 1
	}

	result, err := db.conn.Exec(`
		INSERT INTO news_topics (name, description, display_order, is_active, stories_per_refresh, refresh_interval_minutes, summary_min_words, summary_max_words, ai_provider, is_niche)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Name, t.Description, nextOrder, boolToInt(t.IsActive),
		t.StoriesPerRefresh, t.RefreshIntervalMinutes,
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

func (db *DB) UpdateNewsTopic(t *models.NewsTopic) error {
	_, err := db.conn.Exec(`
		UPDATE news_topics SET name = ?, description = ?, is_active = ?,
		       stories_per_refresh = ?, refresh_interval_minutes = ?,
		       summary_min_words = ?, summary_max_words = ?,
		       ai_provider = ?, is_niche = ?,
		       updated_at = datetime('now')
		WHERE id = ?`,
		t.Name, t.Description, boolToInt(t.IsActive),
		t.StoriesPerRefresh, t.RefreshIntervalMinutes,
		t.SummaryMinWords, t.SummaryMaxWords,
		t.AIProvider, boolToInt(t.IsNiche), t.ID)
	return err
}

func (db *DB) DeleteNewsTopic(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM news_topics WHERE id = ?`, id)
	return err
}

func (db *DB) ToggleNewsTopicActive(id int64, active bool) error {
	_, err := db.conn.Exec(`UPDATE news_topics SET is_active = ?, updated_at = datetime('now') WHERE id = ?`,
		boolToInt(active), id)
	return err
}

func (db *DB) ReorderNewsTopics(ids []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE news_topics SET display_order = ?, updated_at = datetime('now') WHERE id = ?`)
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

func (db *DB) UpdateNewsTopicRefreshTime(id int64) error {
	_, err := db.conn.Exec(`UPDATE news_topics SET last_refreshed_at = datetime('now'), updated_at = datetime('now') WHERE id = ?`, id)
	return err
}

func (db *DB) NewsTopicsDueForRefresh() ([]models.NewsTopic, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, display_order, is_active, stories_per_refresh,
		       refresh_interval_minutes, summary_min_words, summary_max_words,
		       ai_provider, is_niche, last_refreshed_at, created_at, updated_at
		FROM news_topics
		WHERE is_active = 1
		  AND (last_refreshed_at IS NULL
		       OR datetime('now') > datetime(last_refreshed_at, '+' || refresh_interval_minutes || ' minutes'))
		ORDER BY last_refreshed_at ASC NULLS FIRST`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNewsTopics(rows)
}

func scanNewsTopics(rows *sql.Rows) ([]models.NewsTopic, error) {
	var topics []models.NewsTopic
	for rows.Next() {
		var t models.NewsTopic
		var lastRefreshed sql.NullString
		var createdAt, updatedAt string

		if err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.DisplayOrder, &t.IsActive,
			&t.StoriesPerRefresh, &t.RefreshIntervalMinutes,
			&t.SummaryMinWords, &t.SummaryMaxWords,
			&t.AIProvider, &t.IsNiche, &lastRefreshed,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan news topic: %w", err)
		}

		t.CreatedAt, _ = parseTime(createdAt)
		t.UpdatedAt, _ = parseTime(updatedAt)
		if lastRefreshed.Valid {
			parsed, _ := parseTime(lastRefreshed.String)
			t.LastRefreshedAt = &parsed
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

// --- News Sources ---

func (db *DB) GetSourcesForNewsTopic(newsTopicID int64) ([]models.NewsSource, error) {
	rows, err := db.conn.Query(`
		SELECT id, news_topic_id, url, name, is_manual, is_active, failure_count, last_error, created_at
		FROM news_sources WHERE news_topic_id = ? ORDER BY is_manual DESC, id ASC`, newsTopicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNewsSources(rows)
}

func (db *DB) GetActiveSourcesForNewsTopic(newsTopicID int64) ([]models.NewsSource, error) {
	rows, err := db.conn.Query(`
		SELECT id, news_topic_id, url, name, is_manual, is_active, failure_count, last_error, created_at
		FROM news_sources WHERE news_topic_id = ? AND is_active = 1 ORDER BY id ASC`, newsTopicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNewsSources(rows)
}

func (db *DB) AddNewsSource(newsTopicID int64, url, name string, isManual bool) (int64, error) {
	result, err := db.conn.Exec(`
		INSERT INTO news_sources (news_topic_id, url, name, is_manual) VALUES (?, ?, ?, ?)`,
		newsTopicID, url, name, boolToInt(isManual))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) DeleteNewsSource(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM news_sources WHERE id = ?`, id)
	return err
}

func (db *DB) UpdateNewsSourceStatus(id int64, isActive bool, failureCount int, lastError string) error {
	_, err := db.conn.Exec(`
		UPDATE news_sources SET is_active = ?, failure_count = ?, last_error = ? WHERE id = ?`,
		boolToInt(isActive), failureCount, lastError, id)
	return err
}

func (db *DB) ClearAINewsSourcesForTopic(newsTopicID int64) error {
	_, err := db.conn.Exec(`DELETE FROM news_sources WHERE news_topic_id = ? AND is_manual = 0`, newsTopicID)
	return err
}

func scanNewsSources(rows *sql.Rows) ([]models.NewsSource, error) {
	var sources []models.NewsSource
	for rows.Next() {
		var s models.NewsSource
		var createdAt string

		if err := rows.Scan(
			&s.ID, &s.NewsTopicID, &s.URL, &s.Name, &s.IsManual,
			&s.IsActive, &s.FailureCount, &s.LastError, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan news source: %w", err)
		}

		s.CreatedAt, _ = parseTime(createdAt)
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

// --- Stories ---

func (db *DB) ListStoriesByNewsTopic(newsTopicID int64, limit int) ([]models.Story, error) {
	rows, err := db.conn.Query(`
		SELECT id, news_topic_id, title, summary, source_url, source_title, ai_provider, ai_model, published_at, created_at
		FROM stories WHERE news_topic_id = ?
		ORDER BY created_at DESC LIMIT ?`, newsTopicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStories(rows)
}

func (db *DB) CreateStory(s *models.Story) error {
	result, err := db.conn.Exec(`
		INSERT INTO stories (news_topic_id, title, summary, source_url, source_title, ai_provider, ai_model, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		s.NewsTopicID, s.Title, s.Summary, s.SourceURL, s.SourceTitle, s.AIProvider, s.AIModel)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = id
	return nil
}

func (db *DB) DeleteOldStories(newsTopicID int64, keepCount int) error {
	_, err := db.conn.Exec(`
		DELETE FROM stories WHERE news_topic_id = ? AND id NOT IN (
			SELECT id FROM stories WHERE news_topic_id = ? ORDER BY created_at DESC LIMIT ?
		)`, newsTopicID, newsTopicID, keepCount)
	return err
}

func scanStories(rows *sql.Rows) ([]models.Story, error) {
	var stories []models.Story
	for rows.Next() {
		var s models.Story
		var publishedAt, createdAt string

		if err := rows.Scan(
			&s.ID, &s.NewsTopicID, &s.Title, &s.Summary,
			&s.SourceURL, &s.SourceTitle, &s.AIProvider, &s.AIModel,
			&publishedAt, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan story: %w", err)
		}

		s.PublishedAt, _ = parseTime(publishedAt)
		s.CreatedAt, _ = parseTime(createdAt)
		stories = append(stories, s)
	}
	return stories, rows.Err()
}

// --- News Refresh Status ---

func (db *DB) GetNewsRefreshStatus(newsTopicID int64) (*models.NewsRefreshStatus, error) {
	var s models.NewsRefreshStatus
	var lastRefresh, nextRefresh sql.NullString

	err := db.conn.QueryRow(`
		SELECT news_topic_id, last_refresh, next_refresh, status, error_message
		FROM news_refresh_status WHERE news_topic_id = ?`, newsTopicID).Scan(
		&s.NewsTopicID, &lastRefresh, &nextRefresh, &s.Status, &s.ErrorMessage)
	if err != nil {
		return nil, err
	}

	if lastRefresh.Valid {
		s.LastRefresh, _ = parseTime(lastRefresh.String)
	}
	if nextRefresh.Valid {
		s.NextRefresh, _ = parseTime(nextRefresh.String)
	}
	return &s, nil
}

func (db *DB) UpdateNewsRefreshStatus(s *models.NewsRefreshStatus) error {
	_, err := db.conn.Exec(`
		INSERT INTO news_refresh_status (news_topic_id, last_refresh, next_refresh, status, error_message)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(news_topic_id) DO UPDATE SET
			last_refresh = excluded.last_refresh,
			next_refresh = excluded.next_refresh,
			status = excluded.status,
			error_message = excluded.error_message`,
		s.NewsTopicID,
		s.LastRefresh.Format("2006-01-02 15:04:05"),
		s.NextRefresh.Format("2006-01-02 15:04:05"),
		s.Status, s.ErrorMessage)
	return err
}
