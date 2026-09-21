package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/worldline-go/saz/internal/service"
)

// Fault injection for transaction-control failures that real database servers
// cannot reliably produce on demand (especially RELEASE after a successful insert).
type controlFailureConn struct {
	failOn     string
	failure    error
	committed  bool
	rolledBack bool
}

func (c *controlFailureConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c *controlFailureConn) Close() error              { return nil }
func (c *controlFailureConn) Begin() (driver.Tx, error) { return c, nil }
func (c *controlFailureConn) Commit() error             { c.committed = true; return nil }
func (c *controlFailureConn) Rollback() error           { c.rolledBack = true; return nil }
func (c *controlFailureConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.HasPrefix(query, c.failOn) {
		return nil, c.failure
	}
	if c.failOn == "ROLLBACK TO SAVEPOINT" && strings.HasPrefix(query, "INSERT") {
		return nil, errors.New("duplicate key")
	}
	return driver.RowsAffected(1), nil
}

type controlFailureConnector struct{ conn *controlFailureConn }

func (c controlFailureConnector) Connect(context.Context) (driver.Conn, error) { return c.conn, nil }
func (c controlFailureConnector) Driver() driver.Driver                        { return c }
func (c controlFailureConnector) Open(string) (driver.Conn, error)             { return c.conn, nil }

func TestTransferNeverSkipsSavepointFailures(t *testing.T) {
	for _, operation := range []string{"SAVEPOINT", "ROLLBACK TO SAVEPOINT", "RELEASE SAVEPOINT"} {
		t.Run(operation, func(t *testing.T) {
			failure := errors.New("transaction control failed")
			conn := &controlFailureConn{failOn: operation, failure: failure}
			db := sql.OpenDB(controlFailureConnector{conn: conn})
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			d := &Database{DB: map[string]*Info{"target": {DB: db, PlaceHolder: "$"}}}
			result, err := d.IterSet(t.Context(), "target", "target_table", false,
				service.SkipError{Enabled: true}, service.MapType{}, 2,
				[]string{"id"}, transferRows([]any{1}, []any{2}))
			require.ErrorIs(t, err, failure)
			require.Nil(t, result)
			require.False(t, conn.committed)
			require.True(t, conn.rolledBack)
		})
	}
}
