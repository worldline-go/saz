package server

import (
	"context"
	"net/http"

	"github.com/worldline-go/saz/internal/service"
)

func Context(r *http.Request) context.Context {
	user := r.Header.Get("X-User")

	return service.ContextWithUser(r.Context(), user)
}
