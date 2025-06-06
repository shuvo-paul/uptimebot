package repository

import (
	"database/sql"
	"fmt"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/database"
)

type TokenRepository struct {
	db database.Querier
}

func NewTokenRepository(db database.Querier) *TokenRepository {
	return &TokenRepository{db: db}
}

type TokenRepositoryInterface interface {
	SaveToken(token *model.Token) (*model.Token, error)
	GetTokenByValue(token string) (*model.Token, error)
	MarkTokenUsed(tokenID int) error
	GetTokensByUserID(userID int) ([]*model.Token, error)
	InvalidateExistingTokens(userID int, tokenType model.TokenType) error
}

func (r *TokenRepository) SaveToken(token *model.Token) (*model.Token, error) {
	query := `INSERT INTO token (user_id, token, type, expires_at, used) 
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING id`
	err := r.db.QueryRow(query, token.UserID, token.Token, token.Type, token.ExpiresAt, token.Used).Scan(&token.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

func (r *TokenRepository) GetTokenByValue(tokenValue string) (*model.Token, error) {
	var token model.Token
	query := `SELECT id, user_id, token, type, expires_at, used 
			  FROM token WHERE token = $1`
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

func (r *TokenRepository) MarkTokenUsed(tokenID int) error {
	query := `UPDATE token SET used = TRUE WHERE id = $1`
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

func (r *TokenRepository) GetTokensByUserID(userID int) ([]*model.Token, error) {
	query := `SELECT id, user_id, token, type, expires_at, used 
			  FROM token WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*model.Token
	for rows.Next() {
		var token model.Token
		err = rows.Scan(
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

func (r *TokenRepository) InvalidateExistingTokens(userID int, tokenType model.TokenType) error {
	query := `UPDATE token 
			  SET used = TRUE 
			  WHERE user_id = $1 AND type = $2 AND used = FALSE AND expires_at > NOW()`
	_, err := r.db.Exec(query, userID, string(tokenType))
	if err != nil {
		return fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	return nil
}

var _ TokenRepositoryInterface = (*TokenRepository)(nil)
