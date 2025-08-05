package database

import (
	"database/sql"
	"time"
)

type Result struct {
	duration  time.Duration
	sqlResult sql.Result
	rows      []map[string]any
}

func (r *Result) RowsAffected() (int64, error) {
	return r.sqlResult.RowsAffected()
}

func (r *Result) Duration() time.Duration {
	return r.duration
}

func (r *Result) Rows() []map[string]any {
	return r.rows
}
