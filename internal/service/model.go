package service

import "context"

type Note struct {
	ID string `json:"id"`
}

type Storer interface {
	Get(ctx context.Context, id string) (*Note, error)
}

// /////////////////////////////////////////////

type Result interface {
	RowsAffected() (int64, error)
}

type Database interface {
	Run(ctx context.Context, query string, args ...interface{}) (Result, error)
}
