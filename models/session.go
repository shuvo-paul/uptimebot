package models

import (
	"time"
)

type Session struct {
	ID        int
	UserID    int
	Token     string // This will store the hashed token
	CreatedAt time.Time
	ExpiresAt time.Time
}
