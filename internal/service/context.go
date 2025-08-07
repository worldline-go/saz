package service

import "context"

type ContextKey string

const (
	UserContextKey ContextKey = "USER"
)

func UserContext(ctx context.Context) string {
	if user, ok := ctx.Value(UserContextKey).(string); ok {
		return user
	}

	return "unknown"
}

func ContextWithUser(ctx context.Context, user string) context.Context {
	if user == "" {
		return ctx
	}

	return context.WithValue(ctx, UserContextKey, user)
}
