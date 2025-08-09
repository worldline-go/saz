package service

import (
	"context"
	"fmt"

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

func (s *Service) Run(ctx context.Context, cell Cell) (Result, error) {
	if cell.DBType == "" || cell.Content == "" {
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
			iterGet, err := s.db.IterGet(ctx, cell.DBType, cell.Content)
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
