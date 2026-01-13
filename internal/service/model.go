package service

import (
	"context"
	"errors"
	"iter"
	"time"

	"github.com/rakunlabs/query"
	"github.com/worldline-go/types"
)

var (
	ErrNotExists  = errors.New("not exist")
	ErrBadRequest = errors.New("bad request")
)

type Note struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Content   Content                `json:"content"`
	UpdatedBy types.Null[string]     `json:"updated_by"`
	CreatedAt types.Null[types.Time] `json:"created_at"`
	UpdatedAt types.Null[types.Time] `json:"updated_at"`
}

type IDName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Content struct {
	Cells []Cell `json:"cells"`
}

type Cell struct {
	ID          string                 `json:"id"`
	DBType      string                 `json:"db_type"`
	Content     string                 `json:"content"`
	Limit       int64                  `json:"limit"`
	Mode        types.Null[Mode]       `json:"mode,omitzero"`
	Description types.Null[string]     `json:"description,omitzero"`
	Collapsed   types.Null[bool]       `json:"collapsed,omitzero"`
	Enabled     types.Null[bool]       `json:"enabled,omitzero"`
	Result      types.Null[bool]       `json:"result,omitzero"`
	Template    Template               `json:"template"`
	Path        types.Null[string]     `json:"path,omitzero"`
	Dependency  types.Null[Dependency] `json:"dependency,omitzero"`
}

type Dependency struct {
	Enabled bool     `json:"enabled"`
	Names   []string `json:"names"`
}

type Template struct {
	Enabled bool `json:"enabled"`
}

type Mode struct {
	Enabled   bool      `json:"enabled"`
	Name      string    `json:"name"`
	DBType    string    `json:"db_type"`
	Table     string    `json:"table"`
	Wipe      bool      `json:"wipe"`
	SkipError SkipError `json:"skip_error"`
	MapType   MapType   `json:"map_type"`
	Batch     int       `json:"batch"`
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
	Type     string      `json:"type"`
	Nullable bool        `json:"nullable"`
	Template EnableValue `json:"template"`
	Encoding Encoding    `json:"encoding"`
}

type Encoding struct {
	Enabled bool   `json:"enabled"`
	Coding  string `json:"coding"`
}

type EnableValue struct {
	Enabled bool   `json:"enabled"`
	Value   string `json:"value"`
}

// /////////////////////////////////////////////

type Process struct {
	ID        string             `json:"id"`
	Status    string             `json:"status"`
	Info      ProcessInfo        `json:"info"`
	User      types.Null[string] `json:"user,omitzero"`
	CreatedAt types.Time         `json:"created_at"`
	UpdatedAt types.Time         `json:"updated_at"`
}

type ProcessInfo struct {
	Note         string `json:"note,omitempty"`
	Query        string `json:"query,omitempty"`
	Description  string `json:"description,omitempty"`
	RowsAffected int64  `json:"rows_affected,omitempty"`
	Error        string `json:"error,omitempty"`
	Duration     string `json:"duration,omitempty"`
}

// /////////////////////////////////////////////

type Storer interface {
	Get(ctx context.Context, id string) (*Note, error)
	GetWithPath(ctx context.Context, path string) (*Note, error)
	GetNotes(ctx context.Context) ([]IDName, error)
	Save(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id string) error

	GetProcess(ctx context.Context, q *query.Query) ([]Process, error)
	SaveProcess(ctx context.Context, process *Process) error
	DeleteProcess(ctx context.Context, q *query.Query) error
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

	IterGet(ctx context.Context, name, query string, mapType MapType) ([]string, iter.Seq2[[]any, error], error)
	IterSet(ctx context.Context, name, table string, wipe bool, skipError SkipError, mapType MapType, batch int, columns []string, rows iter.Seq2[[]any, error]) (Result, error)
}
