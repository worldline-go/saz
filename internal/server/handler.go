package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rakunlabs/ada"
	"github.com/worldline-go/saz/internal/service"
)

type RunRequest struct {
	Name  string `json:"name"`
	Query string `json:"query"`
	Args  []any  `json:"args,omitempty"`
}

func (s *Server) run(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ada.JSON(w, http.StatusBadRequest, Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})

		return
	}

	result, err := s.service.Run(ctx, req.Name, req.Query, req.Args...)
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			ada.JSON(w, http.StatusNotFound, Response{
				Message: "Database not found",
				Error:   err.Error(),
			})
			return
		}

		ada.JSON(w, http.StatusInternalServerError, Response{
			Message: "Failed to execute query",
			Error:   err.Error(),
		})

		return
	}

	ada.JSON(w, http.StatusOK, Response{
		Data: result.Rows(),
	})
}

func (s *Server) databaseList(w http.ResponseWriter, r *http.Request) {
	dbList := s.service.DatabaseList()

	ada.JSON(w, http.StatusOK, Response{
		Data: dbList,
	})
}
