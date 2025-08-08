package database

import (
	"context"
	"fmt"
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

	return &Result{
		sqlResult: result,
		duration:  time.Since(start),
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
