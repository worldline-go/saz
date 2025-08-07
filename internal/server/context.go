package server

import (
	"context"
	"net/http"

	"github.com/worldline-go/saz/internal/service"
)

func UserContext(r *http.Request) context.Context {
	return service.ContextWithUser(r.Context(), r.Header.Get("X-User"))
}
