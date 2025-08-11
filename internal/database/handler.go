package database

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/cast"
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
		columns: []string{"status"},
		rows: [][]any{
			{"success"},
		},
		rowsAffected: rowsAffected,
		duration:     time.Since(start),
	}, nil
}

func (d *Database) Query(ctx context.Context, name, query string, limit int64) (service.Result, error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	start := time.Now()
	rows := [][]any{}
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
		values, err := rowsIter.SliceScan()
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		valuesStr := make([]any, 0, len(values))
		for _, v := range values {
			valuesStr = append(valuesStr, cast.ToString(v))
		}

		rows = append(rows, valuesStr)

		limit--
		if limit == 0 {
			break
		}
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

func (d *Database) IterGet(ctx context.Context, name, query string, mapType service.MapType) (iter.Seq2[map[string]any, error], error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	rowsIter, err := dbConn.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("run query on database %s: %w", name, err)
	}

	columnTypes, err := rowsIter.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("get column types: %w", err)
	}

	dynamicType := reflect.StructOf(GenerateStruct(columnTypes, mapType))

	return func(yield func(map[string]any, error) bool) {
		defer rowsIter.Close()

		for rowsIter.Next() {
			record := reflect.New(dynamicType).Interface()
			if err := rowsIter.StructScan(record); err != nil {
				_ = !yield(nil, fmt.Errorf("scan row: %w", err))
				return
			}

			mapRecord, err := Struct2Map(record, mapType)
			if err != nil {
				_ = !yield(nil, fmt.Errorf("map struct to map: %w", err))
				return
			}

			if !yield(mapRecord, nil) {
				return
			}
		}
		if err := rowsIter.Err(); err != nil {
			_ = !yield(nil, fmt.Errorf("iterate rows: %w", err))
			return
		}
	}, nil
}

func (d *Database) IterSet(ctx context.Context, name, table string, wipe bool, skipError service.SkipError, rows iter.Seq2[map[string]any, error]) (service.Result, error) {
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

	var savePoint string
	if skipError.Enabled {
		savePoint = fmt.Sprintf("savepoint_%s", ulid.Make())
	}

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

		if skipError.Enabled {
			tx.ExecContext(ctx, "SAVEPOINT "+savePoint)
		}

		if _, err := tx.NamedExecContext(ctx, query, row); err != nil {
			if skipError.Enabled {
				if strings.Contains(err.Error(), skipError.Message) {
					tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savePoint)
					continue
				}
			}

			return nil, fmt.Errorf("insert row: %w; query %s, row %v", err, query, row)
		}

		if skipError.Enabled {
			tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savePoint)
		}

		counter++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction on database %s: %w", name, err)
	}

	return &Result{
		columns: []string{"status"},
		rows: [][]any{
			{"success"},
		},
		duration:     time.Since(start),
		rowsAffected: counter,
	}, nil
}
