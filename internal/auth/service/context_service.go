package service

import (
	"context"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
)

type contextKey string

const userKey contextKey = "user"

func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(userKey).(*model.User)
	return user, ok
}
