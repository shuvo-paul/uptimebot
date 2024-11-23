package services

import (
	"context"

	"github.com/shuvo-paul/sitemonitor/internal/app/models"
)

type contextKey string

const userKey contextKey = "user"

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userKey).(*models.User)
	return user, ok
}
