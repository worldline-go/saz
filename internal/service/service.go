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

	logi.Ctx(ctx).Debug("running cell", "db_type", cell.DBType, "content", cell.Content, "result", cell.Result.V)
	if cell.Result.V {
		return s.db.Query(ctx, cell.DBType, cell.Content)
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
