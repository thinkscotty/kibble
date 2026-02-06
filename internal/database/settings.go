package database

import (
	"github.com/thinkscotty/kibble/internal/models"
)

func (db *DB) GetSetting(key string) (string, error) {
	var value string
	err := db.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (db *DB) SetSetting(key, value string) error {
	_, err := db.conn.Exec(`INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, datetime('now'))`,
		key, value)
	return err
}

func (db *DB) GetAllSettings() (map[string]string, error) {
	rows, err := db.conn.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, rows.Err()
}

func (db *DB) LogAPIUsage(log models.APIUsageLog) error {
	_, err := db.conn.Exec(`
		INSERT INTO api_usage_log (topic_id, facts_requested, facts_generated, facts_discarded, tokens_used, error_message)
		VALUES (?, ?, ?, ?, ?, ?)`,
		log.TopicID, log.FactsRequested, log.FactsGenerated, log.FactsDiscarded,
		log.TokensUsed, log.ErrorMessage)
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

	size, _ := db.DatabaseSizeBytes()
	s.DatabaseSizeBytes = size

	return s, nil
}

func (db *DB) RecentAPIUsage(limit int) ([]models.APIUsageLog, error) {
	rows, err := db.conn.Query(`
		SELECT l.id, l.topic_id, COALESCE(t.name, 'Deleted Topic'), l.facts_requested,
		       l.facts_generated, l.facts_discarded, l.tokens_used,
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
			&log.ErrorMessage, &createdAt); err != nil {
			return nil, err
		}
		log.CreatedAt, _ = parseTime(createdAt)
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
