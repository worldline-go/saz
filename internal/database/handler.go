package database

import (
	"context"
	"database/sql"
	"fmt"
	"iter"
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

	result, err := dbConn.DB.ExecContext(ctx, query)
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
	rowsIter, err := dbConn.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("run query on database %s: %w", name, err)
	}
	defer rowsIter.Close()

	columns, err := rowsIter.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	for rowsIter.Next() {
		values, err := ScanSlice(len(columns), rowsIter)
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

func (d *Database) IterGet(ctx context.Context, name, query string, mapType service.MapType) ([]string, iter.Seq2[[]any, error], error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	rowsIter, err := dbConn.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("run query on database %s: %w", name, err)
	}

	columns, err := rowsIter.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("get columns: %w", err)
	}

	var dynamicSlice []any
	var columnsIndex map[string]int

	if mapType.Enabled {
		columnTypes, err := rowsIter.ColumnTypes()
		if err != nil {
			return nil, nil, fmt.Errorf("get column types: %w", err)
		}

		dynamicSlice = GenerateSlice(columnTypes, mapType)

		columnsIndex = make(map[string]int, len(columns))
		for i, col := range columns {
			columnsIndex[col] = i
		}
	}

	return columns, func(yield func([]any, error) bool) {
		defer rowsIter.Close()

		for rowsIter.Next() {
			var sliceRow []any
			if mapType.Enabled {
				sliceRow, err = ScanSliceWithValues(len(columns), rowsIter, dynamicSlice)
				if err != nil {
					_ = !yield(nil, fmt.Errorf("scan row: %w", err))
					return
				}

				if err := Map(columnsIndex, mapType, sliceRow); err != nil {
					_ = !yield(nil, fmt.Errorf("map struct to map: %w", err))
					return
				}
			} else {
				var err error
				sliceRow, err = ScanSlice(len(columns), rowsIter)
				if err != nil {
					_ = !yield(nil, fmt.Errorf("scan row: %w", err))
					return
				}
			}

			if !yield(sliceRow, nil) {
				return
			}
		}
		if err := rowsIter.Err(); err != nil {
			_ = !yield(nil, fmt.Errorf("iterate rows: %w", err))
			return
		}

		if !yield(nil, nil) {
			return
		}
	}, nil
}

func (d *Database) IterSet(ctx context.Context, name, table string, wipe bool, skipError service.SkipError, mapType service.MapType, batchCount int, columns []string, rows iter.Seq2[[]any, error]) (service.Result, error) {
	dbConn, ok := d.DB[name]
	if !ok {
		return nil, fmt.Errorf("database %s; %w", name, service.ErrNotExists)
	}

	if len(table) == 0 || strings.ContainsAny(table, " \t\n\r") {
		return nil, fmt.Errorf("table name is invalid; %w", service.ErrBadRequest)
	}

	start := time.Now()
	tx, err := dbConn.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction on database %s: %w", name, err)
	}

	defer tx.Rollback()

	if wipe {
		if _, err := tx.ExecContext(ctx, "TRUNCATE TABLE "+table); err != nil {
			return nil, fmt.Errorf("truncate table %s: %w", table, err)
		}
	}

	var counter int64

	var savePoint string
	if skipError.Enabled {
		savePoint = fmt.Sprintf("savepoint_%s", ulid.Make())

		// if batchCount > 1 {
		// 	logi.Ctx(ctx).Info("batch size cannot be use with error skip reduced to 1")

		// 	batchCount = 1
		// }
	}

	if batchCount <= 0 {
		batchCount = 1
	}

	queryBuilderFunc := QueryBuilder(table, columns, dbConn.PlaceHolder)
	query := queryBuilderFunc(batchCount)

	var stmt *sql.Stmt
	if batchCount == 1 && !skipError.Enabled {
		var err error
		stmt, err = tx.PrepareContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("prepare statement: %w", err)
		}
		defer stmt.Close()
	}

	batchHolder := NewBatch(batchCount)

	lastTurn := false
	for row, err := range rows {
		if err != nil {
			return nil, fmt.Errorf("iterate rows: %w", err)
		}

		if len(row) == 0 {
			if batchCount == 1 {
				continue
			}

			lastTurn = true
		}

		if batchCount > 1 {
			if len(row) != 0 {
				batchHolder.AddRow(row)
			}

			if !lastTurn && !batchHolder.IsFull() {
				continue
			}

			size := batchHolder.Size()
			if size == 0 {
				continue
			}

			if size != batchCount {
				query = queryBuilderFunc(size)
			}

			row = batchHolder.Rows()
		}

		if skipError.Enabled {
			tx.ExecContext(ctx, "SAVEPOINT "+savePoint)
		}

		var err error
		if stmt != nil {
			_, err = stmt.ExecContext(ctx, row...)
		} else {
			_, err = tx.ExecContext(ctx, query, row...)
		}

		if err != nil {
			if skipError.Enabled {
				if strings.Contains(err.Error(), skipError.Message) {
					tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savePoint)
					batchHolder.Reset()
					continue
				}
			}

			return nil, fmt.Errorf("insert row: %w; query %s, row %v", err, query, row)
		}

		if skipError.Enabled {
			tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savePoint)
		}

		if batchCount == 1 {
			counter++
		} else {
			counter += int64(batchHolder.Size())
			batchHolder.Reset()
		}
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
