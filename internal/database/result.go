package database

import (
	"time"
)

type Result struct {
	columns      []string
	duration     time.Duration
	rowsAffected int64 // This can be set if using sql.Result
	rows         []map[string]any
}

func (r *Result) RowsAffected() int64 {
	return r.rowsAffected
}

func (r *Result) Duration() time.Duration {
	return r.duration
}

func (r *Result) Rows() []map[string]any {
	return r.rows
}

func (r *Result) Columns() []string {
	return r.columns
}
