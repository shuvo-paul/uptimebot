package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
)

type AccountToken struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Token     string    `db:"token"`
	Type      TokenType `db:"type"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
}

// Core validation methods
func (at *AccountToken) IsExpired() bool {
	return time.Now().After(at.ExpiresAt)
}

func (at *AccountToken) IsUsed() bool {
	return at.Used
}

func (at *AccountToken) MarkUsed() {
	at.Used = true
}

func (at *AccountToken) IsValid() bool {
	return !at.IsExpired() && !at.IsUsed()
}

// Type validation
func (at *AccountToken) ValidateType(expectedType TokenType) error {
	if at.Type != expectedType {
		return errors.New("invalid token type")
	}
	return nil
}

// Constructor with validation
func NewEmailVerificationToken(userID int) (*AccountToken, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return &AccountToken{
		UserID:    userID,
		Token:     uuid.New().String(),
		Type:      TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}
