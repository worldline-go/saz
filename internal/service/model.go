package service

import (
	"context"
	"errors"
	"iter"
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
	Path    string  `json:"path"`
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
	Limit       int64              `json:"limit"`
	Mode        types.Null[Mode]   `json:"mode,omitzero"`
	Description types.Null[string] `json:"description,omitzero"`
	Collapsed   types.Null[bool]   `json:"collapsed,omitzero"`
	Enabled     types.Null[bool]   `json:"enabled,omitzero"`
	Result      types.Null[bool]   `json:"result,omitzero"`
}

type Mode struct {
	Enabled   bool      `json:"enabled"`
	Name      string    `json:"name"`
	DBType    string    `json:"db_type"`
	Table     string    `json:"table"`
	Wipe      bool      `json:"wipe"`
	SkipError SkipError `json:"skip_error"`
	MapType   MapType   `json:"map_type"`
}

type SkipError struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

type MapType struct {
	Enabled     bool                          `json:"enabled"`
	Column      map[string]ColumnType         `json:"column"`
	Destination map[string]ColumnTypeTemplate `json:"destination"`
}

type ColumnType struct {
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

type ColumnTypeTemplate struct {
	Type     string   `json:"type"`
	Nullable bool     `json:"nullable"`
	Template Template `json:"template"`
}

type Template struct {
	Enabled bool   `json:"enabled"`
	Value   string `json:"value"`
}

type Storer interface {
	Get(ctx context.Context, id string) (*Note, error)
	GetWithPath(ctx context.Context, path string) (*Note, error)
	Save(ctx context.Context, note *Note) error
	GetNotes(ctx context.Context) ([]IDName, error)
}

// /////////////////////////////////////////////

type Result interface {
	Columns() []string
	Rows() [][]any
	RowsAffected() int64
	Duration() time.Duration
}

type Database interface {
	DatabaseList() []string

	Query(ctx context.Context, name, query string, limit int64) (Result, error)
	Exec(ctx context.Context, name, query string) (Result, error)

	IterGet(ctx context.Context, name, query string, mapType MapType) (iter.Seq2[map[string]any, error], error)
	IterSet(ctx context.Context, name, table string, wipe bool, skipError SkipError, rows iter.Seq2[map[string]any, error]) (Result, error)
}
