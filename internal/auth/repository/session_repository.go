package repository

import (
	"database/sql"
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/database"
)

type SessionRepository struct {
	db database.Querier
}

func NewSessionRepository(db database.Querier) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *model.Session) error {
	query := `INSERT INTO session (user_id, token, created_at, expires_at) 
			  VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, session.UserID, session.Token,
		session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByToken(token string) (*model.Session, error) {
	var session model.Session
	query := `SELECT user_id, token, created_at, expires_at 
			  FROM session WHERE token = $1`
	err := r.db.QueryRow(query, token).Scan(&session.UserID, &session.Token,
		&session.CreatedAt, &session.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

func (r *SessionRepository) Delete(token string) error {
	query := `DELETE FROM session WHERE token = $1`
	result, err := r.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

type SessionRepositoryInterface interface {
	Create(session *model.Session) error
	GetByToken(token string) (*model.Session, error)
	Delete(sessionID string) error
}

// Ensure SessionRepository implements the interface
var _ SessionRepositoryInterface = (*SessionRepository)(nil)
