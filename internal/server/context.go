package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/rakunlabs/ada/middleware/requestid"
	"github.com/rakunlabs/logi"
	"github.com/worldline-go/saz/internal/service"
)

func Context(r *http.Request) context.Context {
	user := r.Header.Get("X-User")
	ctx := logi.WithContext(r.Context(), slog.With(
		slog.String("request_id", r.Header.Get(requestid.HeaderXRequestID)),
		slog.String("user", user),
	))

	return service.ContextWithUser(ctx, user)
}
