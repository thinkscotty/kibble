package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/thinkscotty/kibble/internal/models"
)

// StoredTrigrams holds minimal data for similarity comparison.
type StoredTrigrams struct {
	ID       int64
	Trigrams string
}

func (db *DB) ListFactsByTopic(topicID int64, limit int) ([]models.Fact, error) {
	rows, err := db.conn.Query(`
		SELECT f.id, f.topic_id, f.content, f.trigrams, f.is_custom, f.is_archived,
		       f.source, f.created_at, f.updated_at
		FROM facts f
		WHERE f.topic_id = ? AND f.is_archived = 0
		ORDER BY f.created_at DESC LIMIT ?`, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFacts(rows)
}

func (db *DB) GetFact(id int64) (models.Fact, error) {
	var f models.Fact
	var createdAt, updatedAt string
	err := db.conn.QueryRow(`
		SELECT f.id, f.topic_id, f.content, f.trigrams, f.is_custom, f.is_archived,
		       f.source, f.created_at, f.updated_at
		FROM facts f WHERE f.id = ?`, id).Scan(
		&f.ID, &f.TopicID, &f.Content, &f.Trigrams, &f.IsCustom, &f.IsArchived,
		&f.Source, &createdAt, &updatedAt)
	if err != nil {
		return f, err
	}
	f.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	f.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return f, nil
}

func (db *DB) CreateFact(f *models.Fact) error {
	result, err := db.conn.Exec(`
		INSERT INTO facts (topic_id, content, trigrams, is_custom, source)
		VALUES (?, ?, ?, ?, ?)`,
		f.TopicID, f.Content, f.Trigrams, boolToInt(f.IsCustom), f.Source)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	f.ID = id
	return nil
}

func (db *DB) UpdateFact(f *models.Fact) error {
	_, err := db.conn.Exec(`
		UPDATE facts SET content = ?, trigrams = ?, updated_at = datetime('now')
		WHERE id = ?`, f.Content, f.Trigrams, f.ID)
	return err
}

func (db *DB) DeleteFact(id int64) error {
	_, err := db.conn.Exec(`UPDATE facts SET is_archived = 1, updated_at = datetime('now') WHERE id = ?`, id)
	return err
}

func (db *DB) HardDeleteFact(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM facts WHERE id = ?`, id)
	return err
}

func (db *DB) SearchFacts(query string, topicID *int64) ([]models.Fact, error) {
	var rows *sql.Rows
	var err error

	likeQuery := "%" + query + "%"

	if topicID != nil {
		rows, err = db.conn.Query(`
			SELECT f.id, f.topic_id, f.content, f.trigrams, f.is_custom, f.is_archived,
			       f.source, f.created_at, f.updated_at
			FROM facts f
			WHERE f.is_archived = 0 AND f.topic_id = ? AND f.content LIKE ?
			ORDER BY f.created_at DESC LIMIT 50`, *topicID, likeQuery)
	} else {
		rows, err = db.conn.Query(`
			SELECT f.id, f.topic_id, f.content, f.trigrams, f.is_custom, f.is_archived,
			       f.source, f.created_at, f.updated_at
			FROM facts f
			WHERE f.is_archived = 0 AND f.content LIKE ?
			ORDER BY f.created_at DESC LIMIT 50`, likeQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFacts(rows)
}

func (db *DB) GetFactTrigramsForTopic(topicID int64) ([]StoredTrigrams, error) {
	rows, err := db.conn.Query(`
		SELECT id, trigrams FROM facts
		WHERE topic_id = ? AND is_archived = 0`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StoredTrigrams
	for rows.Next() {
		var st StoredTrigrams
		if err := rows.Scan(&st.ID, &st.Trigrams); err != nil {
			return nil, err
		}
		result = append(result, st)
	}
	return result, rows.Err()
}

func (db *DB) FactCounts() (total, custom, ai int, err error) {
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM facts WHERE is_archived = 0`).Scan(&total)
	if err != nil {
		return
	}
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM facts WHERE is_archived = 0 AND is_custom = 1`).Scan(&custom)
	if err != nil {
		return
	}
	ai = total - custom
	return
}

func scanFacts(rows *sql.Rows) ([]models.Fact, error) {
	var facts []models.Fact
	for rows.Next() {
		var f models.Fact
		var createdAt, updatedAt string
		if err := rows.Scan(
			&f.ID, &f.TopicID, &f.Content, &f.Trigrams, &f.IsCustom, &f.IsArchived,
			&f.Source, &createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan fact: %w", err)
		}
		f.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		f.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		facts = append(facts, f)
	}
	return facts, rows.Err()
}
