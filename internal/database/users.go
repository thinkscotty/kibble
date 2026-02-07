package database

import (
	"fmt"

	"github.com/thinkscotty/kibble/internal/models"
)

// UserCount returns the number of users in the database.
func (db *DB) UserCount() (int, error) {
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// GetUserByUsername retrieves a user by username.
func (db *DB) GetUserByUsername(username string) (models.User, error) {
	var u models.User
	var createdAt string
	err := db.conn.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE username = ?`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &createdAt)
	if err != nil {
		return u, err
	}
	u.CreatedAt, _ = parseTime(createdAt)
	return u, nil
}

// CreateUser inserts a new user record.
func (db *DB) CreateUser(u *models.User) error {
	result, err := db.conn.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		u.Username, u.PasswordHash,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

// CreateSession inserts a new session record.
func (db *DB) CreateSession(sess *models.Session) error {
	_, err := db.conn.Exec(
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, datetime(?))`,
		sess.Token, sess.UserID, sess.ExpiresAt.UTC().Format("2006-01-02 15:04:05"),
	)
	return err
}

// GetSession retrieves a non-expired session by token.
func (db *DB) GetSession(token string) (models.Session, error) {
	var sess models.Session
	var expiresAt, createdAt string
	err := db.conn.QueryRow(
		`SELECT id, token, user_id, expires_at, created_at
		 FROM sessions
		 WHERE token = ? AND expires_at > datetime('now')`,
		token,
	).Scan(&sess.ID, &sess.Token, &sess.UserID, &expiresAt, &createdAt)
	if err != nil {
		return sess, err
	}
	sess.ExpiresAt, _ = parseTime(expiresAt)
	sess.CreatedAt, _ = parseTime(createdAt)
	return sess, nil
}

// DeleteSession removes a specific session (for logout).
func (db *DB) DeleteSession(token string) error {
	_, err := db.conn.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// DeleteExpiredSessions removes all sessions past their expiry.
func (db *DB) DeleteExpiredSessions() (int64, error) {
	result, err := db.conn.Exec(`DELETE FROM sessions WHERE expires_at <= datetime('now')`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
