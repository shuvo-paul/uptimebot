package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/sitemonitor/config"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	ID        int
	UserID    int
	Token     string // This will store the hashed token
	CreatedAt time.Time
	ExpiresAt time.Time
}

func CreateSession(userID int) (*Session, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	// Generate a unique token
	plainToken := uuid.New().String()

	// Hash the token
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	session := &Session{
		UserID:    userID,
		Token:     string(hashedToken),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires after 24 hours
	}

	query := `INSERT INTO sessions (id, user_id, token, created_at, expires_at) VALUES (?, ?, ?, ?, ?)`
	_, err = config.DB.Exec(query, session.ID, session.UserID, session.Token, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func GetSessionByToken(hashedToken string) (*Session, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var session Session
	query := `SELECT id, user_id, token, created_at, expires_at FROM sessions WHERE token = ?`
	err := config.DB.QueryRow(query, hashedToken).Scan(&session.ID, &session.UserID, &session.Token, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.ExpiresAt.Before(time.Now()) {
		// Session has expired
		DeleteSession(session.ID)
		return nil, fmt.Errorf("session has expired")
	}

	return &session, nil
}

func DeleteSession(sessionID int) error {
	if config.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `DELETE FROM sessions WHERE id = ?`
	_, err := config.DB.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
