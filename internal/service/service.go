package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rakunlabs/logi"
	"github.com/worldline-go/saz/internal/render"
)

type Service struct {
	db    Database
	store Storer
}

func New(db Database, store Storer) *Service {
	return &Service{
		db:    db,
		store: store,
	}
}

func (s *Service) Run(ctx context.Context, cell *Cell) (result Result, err error) {
	if cell == nil || cell.DBType == "" || cell.Content == "" {
		return nil, fmt.Errorf("invalid cell; %w", ErrBadRequest)
	}

	logCell := slog.Group("cell",
		slog.String("description", cell.Description.V),
		slog.String("db_type", cell.DBType),
		slog.String("mode", cell.Mode.V.Name),
	)
	logi.Ctx(ctx).Info("running cell", logCell)

	defer func() {
		if err != nil {
			logi.Ctx(ctx).Error("failed to run cell", logCell, slog.String("error", err.Error()))
		} else {
			logi.Ctx(ctx).Info("cell executed successfully",
				logCell,
				slog.Int64("row_affected", result.RowsAffected()),
				slog.String("duration", result.Duration().String()),
			)
		}
	}()

	content := cell.Content
	if cell.Template.Enabled {
		contentRendered, err := render.Execute(content)
		if err != nil {
			return nil, fmt.Errorf("render content: %w", err)
		}

		content = string(contentRendered)
	}

	if cell.Mode.V.Enabled {
		switch cell.Mode.V.Name {
		case "transfer":
			if cell.Mode.V.Table == "" {
				return nil, fmt.Errorf("transfer mode requires a table name; %w", ErrBadRequest)
			}
			columns, iterGet, err := s.db.IterGet(ctx, cell.DBType, content, cell.Mode.V.MapType)
			if err != nil {
				return nil, fmt.Errorf("get iterator: %w", err)
			}

			// TODO: make better handling of iterators
			defer func() {
				for range iterGet {
					return
				}
			}()

			result, err := s.db.IterSet(ctx, cell.Mode.V.DBType, cell.Mode.V.Table, cell.Mode.V.Wipe, cell.Mode.V.SkipError, cell.Mode.V.MapType, cell.Mode.V.Batch, columns, iterGet)
			if err != nil {
				return nil, fmt.Errorf("set iterator: %w", err)
			}

			return result, nil
		default:
			return nil, fmt.Errorf("unsupported mode %s; %w", cell.Mode.V.Name, ErrBadRequest)
		}
	}

	if cell.Result.V {
		return s.db.Query(ctx, cell.DBType, content, cell.Limit)
	}

	return s.db.Exec(ctx, cell.DBType, content)
}

func (s *Service) RunNote(ctx context.Context, notePath string) (err error) {
	if notePath == "" {
		return fmt.Errorf("note path is empty; %w", ErrBadRequest)
	}

	note, err := s.store.GetWithPath(ctx, notePath)
	if err != nil {
		return fmt.Errorf("get note by path %s: %w", notePath, err)
	}

	logNote := slog.Group("note", slog.String("name", note.Name), slog.String("path", note.Path))

	defer func() {
		if err != nil {
			logi.Ctx(ctx).Error("failed to run note", logNote, slog.String("error", err.Error()))
		} else {
			logi.Ctx(ctx).Info("note executed successfully", logNote)
		}
	}()

	logi.Ctx(ctx).Info("starting note execution", logNote)

	for i := range note.Content.Cells {
		logCell := slog.Group("cell", slog.String("description", note.Content.Cells[i].Description.V), slog.Int("number", i+1))
		ctxCell := logi.WithContext(ctx, logi.Ctx(ctx).With(logNote, logCell))
		if !note.Content.Cells[i].Enabled.V {
			logi.Ctx(ctx).Info("cell is disabled, skipping execution", logCell)
			continue
		}

		note.Content.Cells[i].Result.V = false
		_, err := s.Run(ctxCell, &note.Content.Cells[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) RunNoteCell(ctx context.Context, notePath string, cell int) (result Result, err error) {
	if notePath == "" {
		return nil, fmt.Errorf("note path is empty; %w", ErrBadRequest)
	}
	if cell < 1 {
		return nil, fmt.Errorf("cell number is invalid; %w", ErrBadRequest)
	}

	note, err := s.store.GetWithPath(ctx, notePath)
	if err != nil {
		return nil, fmt.Errorf("get note by path %s: %w", notePath, err)
	}

	if cell > len(note.Content.Cells) {
		return nil, fmt.Errorf("cell number %d is out of range; %w", cell, ErrBadRequest)
	}

	logNote := slog.Group("note", slog.String("name", note.Name), slog.String("path", note.Path))
	logCell := slog.Group("cell", slog.String("description", note.Content.Cells[cell-1].Description.V), slog.Int("number", cell))
	ctxCell := logi.WithContext(ctx, logi.Ctx(ctx).With(logNote, logCell))

	defer func() {
		if err != nil {
			logi.Ctx(ctxCell).Error("failed to run cell", logNote, logCell, slog.String("error", err.Error()))
		} else {
			logi.Ctx(ctxCell).Info("cell executed successfully", logNote, logCell)
		}
	}()

	logi.Ctx(ctxCell).Info("starting cell execution", logNote, logCell)

	result, err = s.Run(ctxCell, &note.Content.Cells[cell-1])
	if err != nil {
		return nil, err
	}

	if note.Content.Cells[cell-1].Result.V {
		logi.Ctx(ctxCell).Info("cell result",
			slog.Int64("rows_affected", result.RowsAffected()),
			slog.Int("columns", len(result.Columns())),
			slog.Int("rows", len(result.Rows())),
			slog.String("duration", result.Duration().String()),
		)
	}

	return result, nil
}

func (s *Service) DatabaseList() []string {
	return s.db.DatabaseList()
}

func (s *Service) GetNote(ctx context.Context, id string) (*Note, error) {
	return s.store.Get(ctx, id)
}

func (s *Service) SaveNote(ctx context.Context, note *Note) error {
	if note == nil {
		return fmt.Errorf("note is nil; %w", ErrBadRequest)
	}
	if note.ID == "" || note.Name == "" || note.Path == "" {
		return fmt.Errorf("invalid Name, ID, or Path; %w", ErrBadRequest)
	}

	return s.store.Save(ctx, note)
}

func (s *Service) DeleteNote(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("note ID is empty; %w", ErrBadRequest)
	}

	return s.store.Delete(ctx, id)
}

func (s *Service) GetNotes(ctx context.Context) ([]IDName, error) {
	return s.store.GetNotes(ctx)
}
