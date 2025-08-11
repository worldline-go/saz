package store

import (
	"context"
	"errors"

	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/saz/internal/store/postgres"
)

type StorerClose interface {
	service.Storer
	Close()
}

func New(ctx context.Context, cfg config.Store) (StorerClose, error) {
	var store StorerClose
	var err error
	if cfg.Postgres != nil {
		store, err = postgres.New(ctx, cfg.Postgres)
		if err != nil {
			return nil, err
		}
	}

	if store == nil {
		return nil, errors.New("no store configured")
	}

	return store, nil
}
