package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/worldline-go/saz/internal/service"
)

func Context(r *http.Request) context.Context {
	user := r.Header.Get("X-User")

	return service.ContextWithUser(r.Context(), user)
}

func getValuesFromRequest(r *http.Request) (map[string]any, error) {
	values := make(map[string]any)
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil && errors.Is(err, io.EOF) == false {
		return nil, err
	}

	for k, v := range r.URL.Query() {
		values[k] = v[0]
	}

	return map[string]any{
		"data": values,
	}, nil
}
