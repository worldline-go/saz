package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/rakunlabs/query"
	"github.com/rakunlabs/tummy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/test/container/containerpostgres"
	"github.com/worldline-go/test/utils/dbutils"
)

func (s *PostgresSuite) TestProcessLeasesAndTerminalStates() {
	ctx := s.T().Context()
	store, err := conn(&config.StorePostgres{}, s.container.Sql())
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { _, err := s.container.Sql().Exec("TRUNCATE process"); require.NoError(s.T(), err) })
	first, err := service.New(nil, store, nil)
	require.NoError(s.T(), err)
	second, err := service.New(nil, store, nil)
	require.NoError(s.T(), err)
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pid, err := first.CreateProcess(service.ContextWithUser(ctx, "owner"), service.ProcessInfo{Note: "keep info"}, cancel)
	require.NoError(s.T(), err)
	get := func(id string) service.Process {
		items, err := store.GetProcess(ctx, query.New().AddWhere(query.NewExpressionCmp(query.OperatorEq, "id", id)))
		require.NoError(s.T(), err)
		require.Len(s.T(), items, 1)
		return items[0]
	}
	original := get(pid)
	second.CleanupStaleProcesses(ctx)
	require.Equal(s.T(), service.ProcessStatusRunning, get(pid).Status, "another instance must not fail live work")

	for _, id := range []string{"abandoned", "renewed"} {
		require.NoError(s.T(), store.SaveProcess(ctx, &service.Process{ID: id, Status: service.ProcessStatusRunning, Info: service.ProcessInfo{Note: "keep info"}}))
	}
	_, err = s.container.Sql().ExecContext(ctx, "UPDATE process SET updated_at = NOW() - INTERVAL '5 minutes' WHERE id IN ('abandoned', 'renewed')")
	require.NoError(s.T(), err)
	require.NoError(s.T(), store.HeartbeatProcesses(ctx, []string{"renewed"}))
	second.CleanupStaleProcesses(ctx)
	require.Equal(s.T(), service.ProcessStatusRunning, get("renewed").Status)
	stale := get("abandoned")
	require.Equal(s.T(), service.ProcessStatusFailed, stale.Status)
	require.Equal(s.T(), "process heartbeat expired", stale.Info.Error)
	require.Equal(s.T(), "keep info", stale.Info.Note)
	// A heartbeat or late worker completion must not resurrect an expired lease.
	require.NoError(s.T(), store.HeartbeatProcesses(ctx, []string{"abandoned"}))
	stale.Status = service.ProcessStatusCompleted
	require.ErrorIs(s.T(), store.SaveProcess(ctx, &stale), service.ErrBadRequest)
	require.Equal(s.T(), service.ProcessStatusFailed, get("abandoned").Status)

	// An instance without the local worker cannot claim cancellation succeeded.
	require.ErrorIs(s.T(), second.ActionProcessID(ctx, pid, service.ProcessActionRequest{Action: service.ProcessActionTerminate}), service.ErrBadRequest)
	require.Equal(s.T(), service.ProcessStatusRunning, get(pid).Status)
	require.NoError(s.T(), first.ActionProcessID(ctx, pid, service.ProcessActionRequest{Action: service.ProcessActionTerminate}))
	require.ErrorIs(s.T(), workerCtx.Err(), context.Canceled)
	first.FailProcess(ctx, pid, context.Canceled, time.Second)
	first.CompleteProcess(ctx, pid, 42, time.Second)
	terminated := get(pid)
	require.Equal(s.T(), service.ProcessStatusTerminated, terminated.Status)
	require.Equal(s.T(), original.CreatedAt, terminated.CreatedAt)
	require.Equal(s.T(), original.User, terminated.User)

	// History retention must never remove a still-running process.
	_, err = s.container.Sql().ExecContext(ctx, "UPDATE process SET created_at = NOW() - INTERVAL '1 year'")
	require.NoError(s.T(), err)
	deleted, err := store.DeleteProcessBefore(ctx, time.Now().Add(-time.Hour))
	require.NoError(s.T(), err)
	require.Equal(s.T(), int64(2), deleted)
	require.Equal(s.T(), service.ProcessStatusRunning, get("renewed").Status)
}

type PostgresSuite struct {
	suite.Suite
	container *containerpostgres.Container
}

func (s *PostgresSuite) SetupSuite() {
	tummy.Enable()
	s.container = containerpostgres.New(s.T())
	s.container.ExecuteFolder(s.T(), "./migrations", dbutils.WithValues(map[string]string{
		"table_prefix": "",
	}))
}

func TestExampleTestSuitePostgres(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}

func (s *PostgresSuite) TearDownSuite() {
	s.container.Stop(s.T())
}

func (s *PostgresSuite) Test_Save() {
	postgres, err := conn(&config.StorePostgres{}, s.container.Sql())
	require.NoError(s.T(), err)
	require.NotNil(s.T(), postgres)

	// Test saving a note
	note := &service.Note{
		ID:      "test-note",
		Name:    "Test Note",
		Content: service.Content{Cells: []service.Cell{{ID: "cell1", DBType: "postgres", Content: "SELECT * FROM test"}}},
		Path:    "test-note",
	}

	tummy.Pause()
	tummy.SetTime(tummy.Now().Truncate(time.Microsecond))
	now := tummy.Now()

	err = postgres.Save(service.ContextWithUser(s.T().Context(), "test-user"), note)
	require.NoError(s.T(), err)

	// Verify the note was saved
	savedNote, err := postgres.Get(s.T().Context(), note.ID)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), savedNote)
	require.Equal(s.T(), note.ID, savedNote.ID)
	require.Equal(s.T(), note.Name, savedNote.Name)
	require.Equal(s.T(), note.Content, savedNote.Content)
	require.Equal(s.T(), note.Path, savedNote.Path)
	require.Equal(s.T(), "test-user", savedNote.UpdatedBy.V)

	getCreatedAt := savedNote.CreatedAt.V
	require.Equal(s.T(), getCreatedAt.IsZero(), false, "CreatedAt should not be zero")

	getUpdatedAt := savedNote.UpdatedAt.V
	require.Equal(s.T(), getUpdatedAt.IsZero(), false, "UpdatedAt should not be zero")

	tummy.AddDuration(10 * time.Second)

	// Change the content
	note.Content = service.Content{Cells: []service.Cell{{ID: "celo", DBType: "postgres", Content: "SELECT * FROM test"}}}
	err = postgres.Save(service.ContextWithUser(s.T().Context(), "test-user-2"), note)
	require.NoError(s.T(), err)

	// Get with path
	notePath := "test-note"
	noteByPath, err := postgres.GetWithPath(s.T().Context(), notePath)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), noteByPath)
	require.Equal(s.T(), note.ID, noteByPath.ID)
	require.Equal(s.T(), note.Name, noteByPath.Name)
	require.Equal(s.T(), note.Content, noteByPath.Content)
	require.Equal(s.T(), note.Path, noteByPath.Path)
	require.Equal(s.T(), "test-user-2", noteByPath.UpdatedBy.V)

	// CreatedAt should
	require.Equal(s.T(), now.Truncate(time.Microsecond), noteByPath.CreatedAt.V.Time, "CreatedAt should be the same with now")
	require.Equal(s.T(), getCreatedAt, noteByPath.CreatedAt.V, "CreatedAt should be the same older get")

	// UpdatedAt should be 10 seconds different
	require.Equal(s.T(), noteByPath.UpdatedAt.V.Sub(getUpdatedAt.Time), 10*time.Second, "UpdatedAt should be different")
}
