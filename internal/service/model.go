package service

import (
	"context"
	"errors"
	"time"
)

var ErrNotExists = errors.New("not exist")

type Note struct {
	ID string `json:"id"`
}

type Storer interface {
	Get(ctx context.Context, id string) (*Note, error)
}

// /////////////////////////////////////////////

type Result interface {
	Rows() []map[string]any
	RowsAffected() (int64, error)
	Duration() time.Duration
}

type Database interface {
	DatabaseList() []string

	Run(ctx context.Context, name, query string, args ...any) (Result, error)
}
