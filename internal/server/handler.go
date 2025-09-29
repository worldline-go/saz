package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/rakunlabs/ada"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/render"
	"github.com/worldline-go/saz/internal/service"
)

func (s *Server) run(c *ada.Context) error {
	ctx := context.WithoutCancel(c.Request.Context())

	var cell service.Cell
	if err := json.NewDecoder(c.Request.Body).Decode(&cell); err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	result, err := s.service.Run(ctx, &cell)
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return c.SetStatus(http.StatusNotFound).SendJSON(Response{
				Message: "Resource not found",
				Error:   err.Error(),
			})
		}

		if errors.Is(err, service.ErrBadRequest) {
			return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
				Message: "Invalid cell data",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to execute query",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(ResponseQuery{
		RowsAffected: result.RowsAffected(),
		Columns:      result.Columns(),
		Rows:         result.Rows(),
		Duration:     result.Duration().Truncate(time.Microsecond).String(),
	})
}

func (s *Server) runNote(c *ada.Context) error {
	ctx := context.WithoutCancel(c.Request.Context())

	noteName := c.Request.PathValue("note")

	if err := s.service.RunNote(ctx, noteName); err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return c.SetStatus(http.StatusNotFound).SendJSON(Response{
				Message: "Resource not found",
				Error:   err.Error(),
			})
		}

		if errors.Is(err, service.ErrBadRequest) {
			return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
				Message: "Invalid note name",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to execute query",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Message: "Note executed successfully",
	})
}

func (s *Server) runNoteCell(c *ada.Context) error {
	ctx := context.WithoutCancel(c.Request.Context())

	noteName := c.Request.PathValue("note")
	cellID := c.Request.PathValue("cell")

	cellNumber, err := strconv.Atoi(cellID)
	if err != nil || cellNumber < 1 {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid cell number",
			Error:   "Cell number must be a positive integer",
		})
	}

	result, err := s.service.RunNoteCell(ctx, noteName, cellNumber)
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return c.SetStatus(http.StatusNotFound).SendJSON(Response{
				Message: "Resource not found",
				Error:   err.Error(),
			})
		}

		if errors.Is(err, service.ErrBadRequest) {
			return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
				Message: "Invalid note or cell",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to execute query",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(ResponseQuery{
		RowsAffected: result.RowsAffected(),
		Columns:      result.Columns(),
		Rows:         result.Rows(),
		Duration:     result.Duration().Truncate(time.Microsecond).String(),
	})
}

func (s *Server) info(c *ada.Context) error {
	dbList := s.service.DatabaseList()

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Data: Info{
			Databases: dbList,
			Version:   config.ServerVersion,
		},
	})
}

func (s *Server) putNote(c *ada.Context) error {
	var note service.Note
	if err := json.NewDecoder(c.Request.Body).Decode(&note); err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid note format",
			Error:   err.Error(),
		})
	}

	note.ID = c.Request.PathValue("id")

	if err := s.service.SaveNote(c.Request.Context(), &note); err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
				Message: "Invalid note data",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to save note",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Message: "Note saved successfully",
	})
}

func (s *Server) deleteNote(c *ada.Context) error {
	ctx := c.Request.Context()
	id := c.Request.PathValue("id")
	if id == "" {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Note ID is required",
		})
	}

	if err := s.service.DeleteNote(ctx, id); err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return c.SetStatus(http.StatusNotFound).SendJSON(Response{
				Message: "Note not found",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to delete note",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Message: "Note deleted successfully",
	})
}

func (s *Server) getNotes(c *ada.Context) error {
	ctx := c.Request.Context()
	notes, err := s.service.GetNotes(ctx)
	if err != nil {
		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to retrieve notes",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Data: notes,
	})
}

func (s *Server) getNote(c *ada.Context) error {
	ctx := c.Request.Context()
	id := c.Request.PathValue("id")
	if id == "" {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Note ID is required",
		})
	}

	note, err := s.service.GetNote(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return c.SetStatus(http.StatusNotFound).SendJSON(Response{
				Message: "Note not found",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to retrieve note",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Data: note,
	})
}

func (s *Server) render(c *ada.Context) error {
	var req RenderRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	data, err := render.ExecuteWithData(req.Content, req.Data)
	if err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Failed to render data",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Data: string(data),
	})
}
