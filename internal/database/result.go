package database

import (
	"database/sql"
	"time"
)

type Result struct {
	columns   []string
	duration  time.Duration
	sqlResult sql.Result
	rows      []map[string]any
}

func (r *Result) RowsAffected() int64 {
	// v, _ := r.sqlResult.RowsAffected()
	// return v
	return 0
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
