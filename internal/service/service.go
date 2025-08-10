package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rakunlabs/logi"
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

func (s *Service) Run(ctx context.Context, cell *Cell) (Result, error) {
	if cell == nil || cell.DBType == "" || cell.Content == "" {
		return nil, fmt.Errorf("invalid cell; %w", ErrBadRequest)
	}

	logi.Ctx(ctx).Debug("running cell",
		"db_type", cell.DBType,
		"content", cell.Content,
		"result", cell.Result.V,
		"mode", cell.Mode.V,
	)

	if cell.Mode.V.Enabled {
		switch cell.Mode.V.Name {
		case "transfer":
			if cell.Mode.V.Table == "" {
				return nil, fmt.Errorf("transfer mode requires a table name; %w", ErrBadRequest)
			}
			iterGet, err := s.db.IterGet(ctx, cell.DBType, cell.Content, cell.Mode.V.MapType)
			if err != nil {
				return nil, fmt.Errorf("get iterator: %w", err)
			}

			// TODO: make better handling of iterators
			defer func() {
				for range iterGet {
					return
				}
			}()

			result, err := s.db.IterSet(ctx, cell.Mode.V.DBType, cell.Mode.V.Table, cell.Mode.V.Wipe, iterGet)
			if err != nil {
				return nil, fmt.Errorf("set iterator: %w", err)
			}

			return result, nil
		default:
			return nil, fmt.Errorf("unsupported mode %s; %w", cell.Mode.V.Name, ErrBadRequest)
		}
	}

	if cell.Result.V {
		return s.db.Query(ctx, cell.DBType, cell.Content, cell.Limit)
	}

	return s.db.Exec(ctx, cell.DBType, cell.Content)
}

func (s *Service) RunNote(ctx context.Context, notePath string) error {
	if notePath == "" {
		return fmt.Errorf("note path is empty; %w", ErrBadRequest)
	}

	note, err := s.store.GetWithPath(ctx, notePath)
	if err != nil {
		return fmt.Errorf("get note by path %s: %w", notePath, err)
	}

	logNote := slog.Group("note", slog.String("name", note.Name), slog.String("path", note.Path))
	logi.Ctx(ctx).Info("starting note execution", logNote)

	for i := range note.Content.Cells {
		logCell := slog.Group("cell", slog.String("description", note.Content.Cells[i].Description.V), slog.Int("number", i+1))
		logi.Ctx(ctx).Info("executing cell", logNote, logCell)

		note.Content.Cells[i].Result.V = false
		result, err := s.Run(ctx, &note.Content.Cells[i])
		if err != nil {
			logi.Ctx(ctx).Error("failed to run cell", logNote, logCell, slog.String("error", err.Error()))

			return err
		}

		logi.Ctx(ctx).Info("cell executed successfully",
			logNote, logCell,
			slog.Int64("row_affected", result.RowsAffected()),
			slog.Duration("duration", result.Duration()),
		)
	}

	return nil
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

func (s *Service) GetNotes(ctx context.Context) ([]IDName, error) {
	return s.store.GetNotes(ctx)
}
