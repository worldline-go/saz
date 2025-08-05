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

func (d *Database) Run(ctx context.Context, name, query string, args ...any) (service.Result, error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	start := time.Now()
	rows := []map[string]any{}
	rowsIter, err := dbConn.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query on database %s: %w", name, err)
	}
	defer rowsIter.Close()

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
		rows:     rows,
		duration: time.Since(start),
	}, nil
}
