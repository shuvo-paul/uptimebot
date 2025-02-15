package repository

import (
	"database/sql"
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
)

type VerificationTokenRepository struct {
	db *sql.DB
}

func NewVerificationTokenRepository(db *sql.DB) *VerificationTokenRepository {
	return &VerificationTokenRepository{db: db}
}

type VerificationTokenRepositoryInterface interface {
	SaveToken(token *model.AccountToken) (*model.AccountToken, error)
	GetTokenByValue(token string) (*model.AccountToken, error)
	MarkTokenUsed(tokenID int) error
	GetTokensByUserID(userID int) ([]*model.AccountToken, error)
	InvalidateExistingTokens(userID int, tokenType model.TokenType) error
}

func (r *VerificationTokenRepository) SaveToken(token *model.AccountToken) (*model.AccountToken, error) {
	query := `INSERT INTO account_tokens (user_id, token, type, expires_at, used) 
			  VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query, token.UserID, token.Token, token.Type, token.ExpiresAt, token.Used)
	if err != nil {
		return nil, fmt.Errorf("failed to save verification token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	token.ID = int(id)
	return token, nil
}

func (r *VerificationTokenRepository) GetTokenByValue(tokenValue string) (*model.AccountToken, error) {
	var token model.AccountToken
	query := `SELECT id, user_id, token, type, expires_at, used 
			  FROM account_tokens WHERE token = ?`
	err := r.db.QueryRow(query, tokenValue).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.Type,
		&token.ExpiresAt,
		&token.Used,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get verification token: %w", err)
	}
	return &token, nil
}

func (r *VerificationTokenRepository) MarkTokenUsed(tokenID int) error {
	query := `UPDATE account_tokens SET used = TRUE WHERE id = ?`
	result, err := r.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no token found with ID: %d", tokenID)
	}
	return nil
}

func (r *VerificationTokenRepository) GetTokensByUserID(userID int) ([]*model.AccountToken, error) {
	query := `SELECT id, user_id, token, type, expires_at, used 
			  FROM account_tokens WHERE user_id = ?`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*model.AccountToken
	for rows.Next() {
		var token model.AccountToken
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Token,
			&token.Type,
			&token.ExpiresAt,
			&token.Used,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verification token: %w", err)
		}
		tokens = append(tokens, &token)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating verification tokens: %w", err)
	}
	return tokens, nil
}

func (r *VerificationTokenRepository) InvalidateExistingTokens(userID int, tokenType model.TokenType) error {
	query := `UPDATE account_tokens 
			  SET used = TRUE 
			  WHERE user_id = ? AND type = ? AND used = FALSE AND expires_at > datetime('now')`
	_, err := r.db.Exec(query, userID, string(tokenType))
	if err != nil {
		return fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	return nil
}

var _ VerificationTokenRepositoryInterface = (*VerificationTokenRepository)(nil)
