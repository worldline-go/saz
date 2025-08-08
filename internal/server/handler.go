package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rakunlabs/ada"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
)

func (s *Server) run(w http.ResponseWriter, r *http.Request) {
	ctx := Context(r)

	var req service.Cell
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ada.JSON(w, http.StatusBadRequest, Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})

		return
	}

	result, err := s.service.Run(ctx, req)
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
		RowsAffected: result.RowsAffected(),
		Columns:      result.Columns(),
		Data:         result.Rows(),
	})
}

func (s *Server) info(w http.ResponseWriter, r *http.Request) {
	dbList := s.service.DatabaseList()

	ada.JSON(w, http.StatusOK, Response{
		Data: Info{
			Databases: dbList,
			Version:   config.ServerVersion,
		},
	})
}

func (s *Server) putNote(w http.ResponseWriter, r *http.Request) {
	var note service.Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		ada.JSON(w, http.StatusBadRequest, Response{
			Message: "Invalid note format",
			Error:   err.Error(),
		})
		return
	}

	note.ID = r.PathValue("id")

	if err := s.service.SaveNote(Context(r), &note); err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			ada.JSON(w, http.StatusBadRequest, Response{
				Message: "Invalid note data",
				Error:   err.Error(),
			})
			return
		}

		ada.JSON(w, http.StatusInternalServerError, Response{
			Message: "Failed to save note",
			Error:   err.Error(),
		})
		return
	}

	ada.JSON(w, http.StatusOK, Response{
		Message: "Note saved successfully",
	})
}

func (s *Server) getNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	notes, err := s.service.GetNotes(ctx)
	if err != nil {
		ada.JSON(w, http.StatusInternalServerError, Response{
			Message: "Failed to retrieve notes",
			Error:   err.Error(),
		})
		return
	}

	ada.JSON(w, http.StatusOK, Response{
		Data: notes,
	})
}

func (s *Server) getNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		ada.JSON(w, http.StatusBadRequest, Response{
			Message: "Note ID is required",
		})
		return
	}

	note, err := s.service.GetNote(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			ada.JSON(w, http.StatusNotFound, Response{
				Message: "Note not found",
				Error:   err.Error(),
			})
			return
		}

		ada.JSON(w, http.StatusInternalServerError, Response{
			Message: "Failed to retrieve note",
			Error:   err.Error(),
		})
		return
	}

	ada.JSON(w, http.StatusOK, Response{
		Data: note,
	})
}
