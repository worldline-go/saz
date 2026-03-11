package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/rakunlabs/ada"
	"github.com/rakunlabs/query"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/render"
	"github.com/worldline-go/saz/internal/service"
)

type CellWithValues struct {
	service.Cell

	Cells  map[string]*service.Cell `json:"cells"`
	Values map[string]any           `json:"values"`
}

func (s *Server) run(c *ada.Context) error {
	baseCtx := context.WithoutCancel(c.Request.Context())
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	var cell CellWithValues
	if err := json.NewDecoder(c.Request.Body).Decode(&cell); err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Create process record
	pid, err := s.service.CreateProcess(baseCtx, service.ProcessInfo{
		Query:       cell.Cell.Content,
		Description: cell.Cell.Description.V,
	}, cancel)
	if err != nil {
		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to create process",
			Error:   err.Error(),
		})
	}

	start := time.Now()

	cellResult := make(map[string]any)
	for key, depCell := range cell.Cells {
		depResult, err := s.service.Run(ctx, depCell, cell.Values, nil)
		if err != nil {
			s.service.FailProcess(baseCtx, pid, err, time.Since(start))
			return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
				Message: "Failed to execute dependency cell; " + depCell.Description.V,
				Error:   err.Error(),
			})
		}

		depRows := service.DataToMap(depResult.Columns(), depResult.Rows())
		cellResult[key] = depRows
	}

	cell.Values["cells"] = cellResult

	result, err := s.service.Run(ctx, &cell.Cell, cell.Values, nil)
	if err != nil {
		s.service.FailProcess(baseCtx, pid, err, time.Since(start))

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

	s.service.CompleteProcess(baseCtx, pid, result.RowsAffected(), result.Duration())

	return c.SetStatus(http.StatusOK).SendJSON(ResponseQuery{
		RowsAffected: result.RowsAffected(),
		Columns:      result.Columns(),
		Rows:         result.Rows(),
		Duration:     result.Duration().Truncate(time.Microsecond).String(),
	})
}

func (s *Server) runBackground(c *ada.Context) error {
	baseCtx := context.WithoutCancel(c.Request.Context())

	var cell CellWithValues
	if err := json.NewDecoder(c.Request.Body).Decode(&cell); err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Create a context that is NOT tied to the HTTP request lifecycle.
	// The goroutine owns this context; it will be cancelled when the query
	// finishes, fails, or is terminated via the Process page.
	ctx, cancel := context.WithCancel(baseCtx)

	// Create process record — registers the cancel func for termination support.
	pid, err := s.service.CreateProcess(baseCtx, service.ProcessInfo{
		Query:       cell.Cell.Content,
		Description: cell.Cell.Description.V,
	}, cancel)
	if err != nil {
		cancel()
		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to create process",
			Error:   err.Error(),
		})
	}

	// Launch execution in background goroutine
	go func() {
		defer cancel()

		start := time.Now()

		cellResult := make(map[string]any)
		for key, depCell := range cell.Cells {
			depResult, err := s.service.Run(ctx, depCell, cell.Values, nil)
			if err != nil {
				s.service.FailProcess(baseCtx, pid, err, time.Since(start))
				return
			}

			depRows := service.DataToMap(depResult.Columns(), depResult.Rows())
			cellResult[key] = depRows
		}

		cell.Values["cells"] = cellResult

		result, err := s.service.Run(ctx, &cell.Cell, cell.Values, nil)
		if err != nil {
			s.service.FailProcess(baseCtx, pid, err, time.Since(start))
			return
		}

		s.service.CompleteProcess(baseCtx, pid, result.RowsAffected(), result.Duration())
	}()

	return c.SetStatus(http.StatusAccepted).SendJSON(ResponseBackground{
		PID:     pid,
		Message: "Query running in background",
	})
}

func (s *Server) runNote(c *ada.Context) error {
	baseCtx := context.WithoutCancel(c.Request.Context())
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	noteName := c.Request.PathValue("note")

	values, err := getValuesFromRequest(c.Request)
	if err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Create process record
	pid, err := s.service.CreateProcess(baseCtx, service.ProcessInfo{
		Note: noteName,
	}, cancel)
	if err != nil {
		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to create process",
			Error:   err.Error(),
		})
	}

	start := time.Now()

	cellInfos, err := s.service.RunNote(ctx, noteName, values)
	if err != nil {
		s.service.FailProcessWithCells(baseCtx, pid, err, time.Since(start), cellInfos)

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

	s.service.CompleteProcessWithCells(baseCtx, pid, 0, time.Since(start), cellInfos)

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Message: "Note executed successfully",
	})
}

func (s *Server) runNoteCell(c *ada.Context) error {
	baseCtx := context.WithoutCancel(c.Request.Context())
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	noteName := c.Request.PathValue("note")
	cellPath := c.Request.PathValue("cell")

	values, err := getValuesFromRequest(c.Request)
	if err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Create process record
	pid, err := s.service.CreateProcess(baseCtx, service.ProcessInfo{
		Note:        noteName,
		Description: cellPath,
	}, cancel)
	if err != nil {
		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to create process",
			Error:   err.Error(),
		})
	}

	start := time.Now()

	result, err := s.service.RunNoteCell(ctx, noteName, cellPath, values)
	if err != nil {
		s.service.FailProcess(baseCtx, pid, err, time.Since(start))

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

	s.service.CompleteProcess(baseCtx, pid, result.RowsAffected(), result.Duration())

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

// //////////////////////////////////////////

func (s *Server) getProcess(c *ada.Context) error {
	q, err := query.Parse(c.Request.URL.RawQuery)
	if err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
	}

	processes, err := s.service.GetProcess(c.Request.Context(), q)
	if err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Failed to retrieve processes",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Data: processes,
	})
}

func (s *Server) getProcessID(c *ada.Context) error {
	pid := c.Request.PathValue("pid")
	if pid == "" {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Process ID is required",
		})
	}

	process, err := s.service.GetProcessID(c.Request.Context(), pid)
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return c.SetStatus(http.StatusNotFound).SendJSON(Response{
				Message: "Process not found",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to retrieve process",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Data: process,
	})
}

func (s *Server) actionProcessID(c *ada.Context) error {
	pid := c.Request.PathValue("pid")
	if pid == "" {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Process ID is required",
		})
	}

	var action service.ProcessActionRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&action); err != nil {
		return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	if err := s.service.ActionProcessID(c.Request.Context(), pid, action); err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return c.SetStatus(http.StatusNotFound).SendJSON(Response{
				Message: "Process not found",
				Error:   err.Error(),
			})
		}

		if errors.Is(err, service.ErrBadRequest) {
			return c.SetStatus(http.StatusBadRequest).SendJSON(Response{
				Message: "Invalid action",
				Error:   err.Error(),
			})
		}

		return c.SetStatus(http.StatusInternalServerError).SendJSON(Response{
			Message: "Failed to perform action on process",
			Error:   err.Error(),
		})
	}

	return c.SetStatus(http.StatusOK).SendJSON(Response{
		Message: "Action performed successfully",
	})
}

// //////////////////////////////////////////
