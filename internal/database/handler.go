package database

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/worldline-go/saz/internal/service"
)

func (d *Database) DatabaseList() []string {
	dbList := make([]string, 0, len(d.DB))
	for name := range d.DB {
		dbList = append(dbList, name)
	}

	return dbList
}

func (d *Database) Exec(ctx context.Context, name, query string) (service.Result, error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	start := time.Now()
	result, err := dbConn.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("execute query on database %s: %w", name, err)
	}

	rowsAffected, _ := result.RowsAffected()
	return &Result{
		rowsAffected: rowsAffected,
		duration:     time.Since(start),
	}, nil
}

func (d *Database) Query(ctx context.Context, name, query string) (service.Result, error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	start := time.Now()
	rows := []map[string]any{}
	rowsIter, err := dbConn.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("run query on database %s: %w", name, err)
	}
	defer rowsIter.Close()

	columns, err := rowsIter.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	for rowsIter.Next() {
		row := make(map[string]any)
		if err := rowsIter.MapScan(row); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		rows = append(rows, row)
	}
	if err := rowsIter.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return &Result{
		columns:  columns,
		rows:     rows,
		duration: time.Since(start),
	}, nil
}

// /////////////////////////////////////////////

func (d *Database) IterGet(ctx context.Context, name, query string) (iter.Seq2[map[string]any, error], error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	rowsIter, err := dbConn.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("run query on database %s: %w", name, err)
	}

	return func(yield func(map[string]any, error) bool) {
		defer rowsIter.Close()

		for rowsIter.Next() {
			row := make(map[string]any)
			if err := rowsIter.MapScan(row); err != nil {
				_ = !yield(nil, fmt.Errorf("scan row: %w", err))
			}

			if !yield(row, nil) {
				return
			}
		}
		if err := rowsIter.Err(); err != nil {
			_ = !yield(nil, fmt.Errorf("iterate rows: %w", err))
			return
		}
	}, nil
}

func (d *Database) IterSet(ctx context.Context, name, table string, wipe bool, rows iter.Seq2[map[string]any, error]) (service.Result, error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	if len(table) == 0 || strings.ContainsAny(table, " \t\n\r") {
		return nil, fmt.Errorf("table name is invalid; %w", service.ErrBadRequest)
	}

	start := time.Now()
	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction on database %s: %w", name, err)
	}

	defer tx.Rollback()

	if wipe {
		if _, err := tx.Exec("TRUNCATE TABLE " + table); err != nil {
			return nil, fmt.Errorf("truncate table %s: %w", table, err)
		}
	}

	var counter int64

	for row, err := range rows {
		if err != nil {
			return nil, fmt.Errorf("iterate rows: %w", err)
		}

		if len(row) == 0 {
			continue
		}

		// row is a map[string]any add that with using sqlx
		columns := make([]string, 0, len(row))
		placeholders := make([]string, 0, len(row))
		for k := range row {
			columns = append(columns, k)
			placeholders = append(placeholders, ":"+k)
		}
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ","), strings.Join(placeholders, ","))

		if _, err := tx.NamedExec(query, row); err != nil {
			return nil, fmt.Errorf("insert row: %w", err)
		}

		counter++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction on database %s: %w", name, err)
	}

	return &Result{
		columns:      []string{"status"},
		rows:         []map[string]any{{"status": "success"}},
		duration:     time.Since(start),
		rowsAffected: counter,
	}, nil
}
