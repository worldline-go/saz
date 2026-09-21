package database

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strconv"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/test/container/containerpostgres"
)

func transferRows(values ...[]any) iter.Seq2[[]any, error] {
	return func(yield func([]any, error) bool) {
		for _, row := range values {
			if !yield(row, nil) {
				return
			}
		}
	}
}

func (s *DatabaseSuite) TestSkipErrorPreservesValidRows() {
	ctx := s.T().Context()
	_, err := s.container.Sql().ExecContext(ctx, "CREATE TABLE transfer_skip (id integer PRIMARY KEY CHECK (id > 0))")
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { _, err := s.container.Sql().Exec("DROP TABLE transfer_skip"); require.NoError(s.T(), err) })
	for _, batch := range []int{1, 3, 10} {
		s.Run(fmt.Sprintf("batch=%d", batch), func() {
			result, err := s.Database.IterSet(ctx, "postgres", "transfer_skip", true,
				service.SkipError{Enabled: true, Message: "duplicate key"}, service.MapType{}, batch,
				[]string{"id"}, transferRows([]any{1}, []any{2}, []any{1}, []any{3}, []any{4}))
			require.NoError(s.T(), err)
			require.Equal(s.T(), int64(4), result.RowsAffected())
			require.Equal(s.T(), [][]any{{"success", int64(1)}}, result.Rows())
			var ids string
			err = s.container.Sql().QueryRowContext(ctx, "SELECT string_agg(id::text, ',' ORDER BY id) FROM transfer_skip").Scan(&ids)
			require.NoError(s.T(), err)
			require.Equal(s.T(), "1,2,3,4", ids)
		})
	}

	// A different constraint failure during row-by-row retry must roll back
	// the entire transfer, including the initial TRUNCATE and successful rows.
	_, err = s.Database.IterSet(ctx, "postgres", "transfer_skip", true,
		service.SkipError{Enabled: true, Message: "duplicate key"}, service.MapType{}, 4,
		[]string{"id"}, transferRows([]any{5}, []any{5}, []any{-1}, []any{6}))
	require.ErrorContains(s.T(), err, "check constraint")
	var ids string
	err = s.container.Sql().QueryRowContext(ctx, "SELECT string_agg(id::text, ',' ORDER BY id) FROM transfer_skip").Scan(&ids)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "1,2,3,4", ids)
}

func (s *DatabaseSuite) TestTransferControlAndSourceErrorsRollback() {
	ctx := s.T().Context()
	_, err := s.container.Sql().ExecContext(ctx, "CREATE TABLE transfer_errors (id integer); INSERT INTO transfer_errors VALUES (99)")
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { _, err := s.container.Sql().Exec("DROP TABLE transfer_errors"); require.NoError(s.T(), err) })
	for _, mode := range []string{"source", "savepoint"} {
		s.Run(mode, func() {
			transferCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			rows := func(yield func([]any, error) bool) {
				if mode == "savepoint" {
					cancel()
					yield([]any{1}, nil)
					return
				}
				if yield([]any{1}, nil) {
					yield(nil, errors.New("source read failed"))
				}
			}
			_, err := s.Database.IterSet(transferCtx, "postgres", "transfer_errors", true,
				service.SkipError{Enabled: true}, service.MapType{}, 1, []string{"id"}, rows)
			require.Error(s.T(), err)
			if mode == "savepoint" {
				require.ErrorContains(s.T(), err, "create savepoint")
			} else {
				require.ErrorContains(s.T(), err, "source read failed")
			}
			var id int
			err = s.container.Sql().QueryRowContext(ctx, "SELECT id FROM transfer_errors").Scan(&id)
			require.NoError(s.T(), err)
			require.Equal(s.T(), 99, id)
		})
	}
}

type DatabaseSuite struct {
	suite.Suite
	container *containerpostgres.Container
	Database  Database
}

func (s *DatabaseSuite) SetupSuite() {
	s.container = containerpostgres.New(s.T())
	s.container.ExecuteFolder(s.T(), "testdata")

	s.Database = Database{
		DB: map[string]*Info{
			"postgres": {
				DB:          s.container.Sql(),
				PlaceHolder: PlaceHolder("pgx"),
			},
		},
	}
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}

func (s *DatabaseSuite) TearDownSuite() {
	s.container.Stop(s.T())
}

func (s *DatabaseSuite) TearDownTest() {
	_, err := s.container.Sql().ExecContext(s.T().Context(), "TRUNCATE TABLE events")
	require.NoError(s.T(), err)
	_, err = s.container.Sql().ExecContext(s.T().Context(), "TRUNCATE TABLE events_copy")
	require.NoError(s.T(), err)
}

func (s *DatabaseSuite) TestCopyEventsEqualCounts() {
	// add data in the events
	batch := QueryBuilder("events", []string{"id", "name", "created_at"}, s.Database.DB["postgres"].PlaceHolder)

	n := 4
	batchQuery := batch(n)
	var args []any
	for i := range n {
		args = append(args,
			ulid.Make().String(),
			"test_event_"+strconv.Itoa(i),
			"2024-01-01 00:00:00Z",
		)
	}

	_, err := s.container.Sql().ExecContext(s.T().Context(), batchQuery, args...)
	require.NoError(s.T(), err)

	columns, rows, err := s.Database.IterGet(s.T().Context(), "postgres", "select * from events", service.MapType{})
	require.NoError(s.T(), err, "iterGet failed")

	result, err := s.Database.IterSet(s.T().Context(), "postgres", "events_copy", true, service.SkipError{}, service.MapType{}, 2, columns, rows)
	require.NoError(s.T(), err, "iterSet failed")
	require.NotNil(s.T(), result)

	require.Equal(s.T(), int64(4), result.RowsAffected())
}

func (s *DatabaseSuite) TestCopyEventsDiffCounts() {
	// add data in the events
	batch := QueryBuilder("events", []string{"id", "name", "created_at"}, s.Database.DB["postgres"].PlaceHolder)

	n := 11
	batchQuery := batch(n)
	var args []any
	for i := range n {
		args = append(args,
			ulid.Make().String(),
			"test_event_"+strconv.Itoa(i),
			"2024-01-01 00:00:00Z",
		)
	}

	_, err := s.container.Sql().ExecContext(s.T().Context(), batchQuery, args...)
	require.NoError(s.T(), err)

	columns, rows, err := s.Database.IterGet(s.T().Context(), "postgres", "select * from events", service.MapType{})
	require.NoError(s.T(), err, "iterGet failed")

	result, err := s.Database.IterSet(s.T().Context(), "postgres", "events_copy", true, service.SkipError{}, service.MapType{}, 3, columns, rows)
	require.NoError(s.T(), err, "iterSet failed")
	require.NotNil(s.T(), result)

	require.Equal(s.T(), int64(11), result.RowsAffected())
}

func (s *DatabaseSuite) TestTransferPreservesComputedAndNullableValues() {
	ctx := s.T().Context()
	_, err := s.container.Sql().ExecContext(ctx, `CREATE TABLE transfer_values (
		id bigint, amount numeric, label text, active boolean, created_at timestamptz
	)`)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		_, err := s.container.Sql().Exec("DROP TABLE transfer_values")
		require.NoError(s.T(), err)
	})

	const source = `SELECT id,
		COALESCE(amount, id * 100.25) AS amount, label, active, created_at
		FROM (VALUES
			(1::bigint, 11.125::numeric, 'first'::text, true, '2024-01-01Z'::timestamptz),
			(2, NULL, NULL, false, NULL),
			(3, 33.375, 'third', true, '2024-03-01Z'::timestamptz),
			(4, NULL, 'fourth', false, '2024-04-01Z'::timestamptz),
			(5, 12345678901234567890.1234567890123456789, NULL, true, NULL)
		) AS src(id, amount, label, active, created_at) ORDER BY id`
	var expected string
	err = s.container.Sql().QueryRowContext(ctx,
		"SELECT jsonb_agg(to_jsonb(src) ORDER BY id)::text FROM ("+source+") src").Scan(&expected)
	require.NoError(s.T(), err)

	for _, enabled := range []bool{false, true} {
		for _, batchSize := range []int{1, 2, 3, 10} {
			s.Run(fmt.Sprintf("mapping=%t/batch=%d", enabled, batchSize), func() {
				mapping := service.MapType{
					Enabled: enabled,
					Column: map[string]service.ColumnType{
						"amount":     {Type: "number"},
						"label":      {Type: "string", Nullable: true},
						"created_at": {Type: "date", Nullable: true},
					},
				}
				columns, rows, err := s.Database.IterGet(ctx, "postgres", source, mapping)
				require.NoError(s.T(), err)
				result, err := s.Database.IterSet(ctx, "postgres", "transfer_values", true,
					service.SkipError{}, mapping, batchSize, columns, rows)
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(5), result.RowsAffected())
				var actual string
				err = s.container.Sql().QueryRowContext(ctx,
					"SELECT jsonb_agg(to_jsonb(dst) ORDER BY id)::text FROM transfer_values dst").Scan(&actual)
				require.NoError(s.T(), err)
				// PostgreSQL compares numeric values exactly, ignoring trailing zeros.
				// Decoding JSON into float64 here would hide precision loss.
				var equal bool
				err = s.container.Sql().QueryRowContext(ctx,
					"SELECT $1::jsonb = $2::jsonb", expected, actual).Scan(&equal)
				require.NoError(s.T(), err)
				require.True(s.T(), equal, "expected: %s\nactual: %s", expected, actual)
			})
		}
	}
}
