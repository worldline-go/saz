package service

import (
	"context"
	"errors"
	"time"

	"github.com/worldline-go/types"
)

var (
	ErrNotExists  = errors.New("not exist")
	ErrBadRequest = errors.New("bad request")
)

type Note struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Content Content `json:"content"`
}

type IDName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Content struct {
	Cells []Cell `json:"cells"`
}

type Cell struct {
	ID          string             `json:"id"`
	DBType      string             `json:"db_type"`
	Content     string             `json:"content"`
	Description types.Null[string] `json:"description,omitzero"`
	Collapsed   types.Null[bool]   `json:"collapsed,omitzero"`
}

type Storer interface {
	Get(ctx context.Context, id string) (*Note, error)
	Save(ctx context.Context, note *Note) error
	GetNotes(ctx context.Context) ([]IDName, error)
}

// /////////////////////////////////////////////

type Result interface {
	Columns() []string
	Rows() []map[string]any
	RowsAffected() int64
	Duration() time.Duration
}

type Database interface {
	DatabaseList() []string

	Run(ctx context.Context, name, query string, args ...any) (Result, error)
}
