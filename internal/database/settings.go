package database

import (
	"fmt"

	"github.com/thinkscotty/kibble/internal/models"
)

// loadSettingsCache populates the in-memory settings cache from the database.
func (db *DB) loadSettingsCache() error {
	rows, err := db.conn.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return err
	}
	defer rows.Close()

	db.cacheMu.Lock()
	defer db.cacheMu.Unlock()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return err
		}
		db.settings[key] = value
	}
	return rows.Err()
}

func (db *DB) GetSetting(key string) (string, error) {
	db.cacheMu.RLock()
	v, ok := db.settings[key]
	db.cacheMu.RUnlock()
	if ok {
		return v, nil
	}
	// Fallback to DB for keys not yet cached
	var value string
	err := db.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err != nil {
		return "", err
	}
	db.cacheMu.Lock()
	db.settings[key] = value
	db.cacheMu.Unlock()
	return value, nil
}

func (db *DB) SetSetting(key, value string) error {
	_, err := db.conn.Exec(`INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, datetime('now'))`,
		key, value)
	if err != nil {
		return err
	}
	db.cacheMu.Lock()
	db.settings[key] = value
	db.cacheMu.Unlock()
	return nil
}

func (db *DB) GetAllSettings() (map[string]string, error) {
	db.cacheMu.RLock()
	defer db.cacheMu.RUnlock()
	result := make(map[string]string, len(db.settings))
	for k, v := range db.settings {
		result[k] = v
	}
	return result, nil
}

func (db *DB) LogAPIUsage(log models.APIUsageLog) error {
	_, err := db.conn.Exec(`
		INSERT INTO api_usage_log (topic_id, facts_requested, facts_generated, facts_discarded, tokens_used, ai_provider, ai_model, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		log.TopicID, log.FactsRequested, log.FactsGenerated, log.FactsDiscarded,
		log.TokensUsed, log.AIProvider, log.AIModel, log.ErrorMessage)
	return err
}

func (db *DB) GetStats() (models.Stats, error) {
	var s models.Stats

	total, active, err := db.TopicCount()
	if err != nil {
		return s, err
	}
	s.TotalTopics = total
	s.ActiveTopics = active

	totalFacts, custom, ai, err := db.FactCounts()
	if err != nil {
		return s, err
	}
	s.TotalFacts = totalFacts
	s.CustomFacts = custom
	s.AIGeneratedFacts = ai

	db.conn.QueryRow(`SELECT COUNT(*) FROM api_usage_log`).Scan(&s.TotalAPIRequests)
	db.conn.QueryRow(`SELECT COALESCE(SUM(tokens_used), 0) FROM api_usage_log`).Scan(&s.TotalTokensUsed)
	db.conn.QueryRow(`SELECT COALESCE(SUM(facts_discarded), 0) FROM api_usage_log`).Scan(&s.FactsDiscarded)

	// News / Updates stats
	db.conn.QueryRow(`SELECT COUNT(*) FROM news_topics`).Scan(&s.TotalNewsTopics)
	db.conn.QueryRow(`SELECT COUNT(*) FROM news_topics WHERE is_active = 1`).Scan(&s.ActiveNewsTopics)
	db.conn.QueryRow(`SELECT COUNT(*) FROM stories`).Scan(&s.TotalStories)
	db.conn.QueryRow(`SELECT COUNT(*) FROM news_sources`).Scan(&s.TotalNewsSources)
	db.conn.QueryRow(`SELECT COUNT(*) FROM news_sources WHERE is_active = 1`).Scan(&s.ActiveNewsSources)

	size, _ := db.DatabaseSizeBytes()
	s.DatabaseSizeBytes = size

	return s, nil
}

func (db *DB) RecentAPIUsage(limit int) ([]models.APIUsageLog, error) {
	rows, err := db.conn.Query(`
		SELECT l.id, l.topic_id, COALESCE(t.name, 'Deleted Topic'), l.facts_requested,
		       l.facts_generated, l.facts_discarded, l.tokens_used,
		       l.ai_provider, l.ai_model,
		       COALESCE(l.error_message, ''), l.created_at
		FROM api_usage_log l
		LEFT JOIN topics t ON l.topic_id = t.id
		ORDER BY l.created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.APIUsageLog
	for rows.Next() {
		var log models.APIUsageLog
		var createdAt string
		if err := rows.Scan(&log.ID, &log.TopicID, &log.TopicName, &log.FactsRequested,
			&log.FactsGenerated, &log.FactsDiscarded, &log.TokensUsed,
			&log.AIProvider, &log.AIModel,
			&log.ErrorMessage, &createdAt); err != nil {
			return nil, err
		}
		log.CreatedAt, _ = parseTime(createdAt)
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// LogRefresh records a refresh attempt (facts or news) in the refresh_log table.
func (db *DB) LogRefresh(entry models.RefreshLog) error {
	_, err := db.conn.Exec(`
		INSERT INTO refresh_log (topic_type, topic_id, topic_name, status, error_type, error_message, duration_ms, ai_provider, ai_model, item_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.TopicType, entry.TopicID, entry.TopicName, entry.Status,
		entry.ErrorType, entry.ErrorMessage, entry.DurationMs,
		entry.AIProvider, entry.AIModel, entry.ItemCount)
	return err
}

// RecentRefreshLogs returns the N most recent refresh log entries.
func (db *DB) RecentRefreshLogs(limit int) ([]models.RefreshLog, error) {
	rows, err := db.conn.Query(`
		SELECT id, topic_type, topic_id, topic_name, status, error_type, error_message,
		       duration_ms, ai_provider, ai_model, item_count, created_at
		FROM refresh_log
		ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.RefreshLog
	for rows.Next() {
		var entry models.RefreshLog
		var createdAt string
		if err := rows.Scan(&entry.ID, &entry.TopicType, &entry.TopicID, &entry.TopicName,
			&entry.Status, &entry.ErrorType, &entry.ErrorMessage,
			&entry.DurationMs, &entry.AIProvider, &entry.AIModel,
			&entry.ItemCount, &createdAt); err != nil {
			return nil, err
		}
		entry.CreatedAt, _ = parseTime(createdAt)
		logs = append(logs, entry)
	}
	return logs, rows.Err()
}

// CleanOldRefreshLogs removes refresh log entries older than the given number of days.
func (db *DB) CleanOldRefreshLogs(days int) error {
	_, err := db.conn.Exec(`DELETE FROM refresh_log WHERE created_at < datetime('now', ?)`,
		fmt.Sprintf("-%d days", days))
	return err
}
