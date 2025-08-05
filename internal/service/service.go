package service

import "context"

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

func (s *Service) Run(ctx context.Context, name, query string, args ...any) (Result, error) {
	return s.db.Run(ctx, name, query, args...)
}
