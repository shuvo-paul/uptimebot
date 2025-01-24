package model

import (
	"time"
)

type Session struct {
	UserID    int
	Token     string // This will store the hashed token
	CreatedAt time.Time
	ExpiresAt time.Time
}
