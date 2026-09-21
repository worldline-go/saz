package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rakunlabs/query"
	"github.com/stretchr/testify/require"
)

type testStore struct {
	Storer
	note      *Note
	process   Process
	saveErr   error
	heartbeat func(context.Context, []string) error
	expire    func(context.Context, time.Duration) (int64, error)
}

func (s *testStore) GetWithPath(context.Context, string) (*Note, error) { return s.note, nil }
func (s *testStore) GetProcess(context.Context, *query.Query) ([]Process, error) {
	return []Process{s.process}, nil
}
func (s *testStore) SaveProcess(_ context.Context, p *Process) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.process = *p
	return nil
}
func (s *testStore) HeartbeatProcesses(ctx context.Context, ids []string) error {
	return s.heartbeat(ctx, ids)
}
func (s *testStore) FailStaleProcesses(ctx context.Context, age time.Duration) (int64, error) {
	return s.expire(ctx, age)
}

type testDatabase struct {
	Database
	execErr error
}

func (d *testDatabase) Exec(context.Context, string, string) (Result, error) {
	return nil, d.execErr
}

func TestRunNoteCellBounds(t *testing.T) {
	executed := errors.New("reached execution")
	for _, tc := range []struct {
		name  string
		cells []Cell
		index string
		want  error
	}{
		{name: "empty", index: "1", want: ErrBadRequest},
		{name: "zero", cells: []Cell{{}}, index: "0", want: ErrBadRequest},
		{name: "negative", cells: []Cell{{}}, index: "-1", want: ErrBadRequest},
		{name: "past end", cells: []Cell{{}}, index: "2", want: ErrBadRequest},
		{name: "overflow", cells: []Cell{{}}, index: "999999999999999999999", want: ErrBadRequest},
		{name: "last cell", cells: []Cell{{DBType: "postgres", Content: "SELECT 1"}}, index: "1", want: executed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := New(&testDatabase{execErr: executed}, &testStore{note: &Note{Content: Content{Cells: tc.cells}}}, nil)
			require.NoError(t, err)
			_, err = s.RunNoteCell(t.Context(), "note", tc.index, nil)
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestCancelProcessDoesNotAcknowledgeFailedSave(t *testing.T) {
	store := &testStore{process: Process{ID: "pid", Status: ProcessStatusRunning}, saveErr: errors.New("store unavailable")}
	s, err := New(nil, store, nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	s.RegisterCancel("pid", cancel)
	err = s.ActionProcessID(t.Context(), "pid", ProcessActionRequest{Action: ProcessActionTerminate})
	require.ErrorIs(t, err, ErrBadRequest)
	require.NoError(t, ctx.Err())
	require.Equal(t, ProcessStatusRunning, store.process.Status)
	// Retain the registration so a subsequent retry can really cancel it.
	store.saveErr = nil
	require.NoError(t, s.ActionProcessID(t.Context(), "pid", ProcessActionRequest{Action: ProcessActionTerminate}))
	require.ErrorIs(t, ctx.Err(), context.Canceled)
	require.Equal(t, ProcessStatusTerminated, store.process.Status)
}

func TestProcessMaintenanceRenewsLocalLeasesBeforeCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var renewed []string
	var expiry time.Duration
	store := &testStore{
		heartbeat: func(_ context.Context, ids []string) error { renewed = ids; return nil },
		expire: func(_ context.Context, age time.Duration) (int64, error) {
			require.Equal(t, []string{"local"}, renewed)
			expiry = age
			cancel()
			return 0, nil
		},
	}
	s, err := New(nil, store, nil)
	require.NoError(t, err)
	s.RegisterCancel("local", func() {})
	s.StartProcessMaintenance(ctx)
	require.Equal(t, processStaleAfter, expiry)
}

func TestProcessMaintenanceDoesNotExpireAfterHeartbeatFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	store := &testStore{
		heartbeat: func(context.Context, []string) error {
			cancel()
			return errors.New("store unavailable")
		},
		expire: func(context.Context, time.Duration) (int64, error) {
			t.Fatal("cleanup must not run after failed heartbeat")
			return 0, nil
		},
	}
	s, err := New(nil, store, nil)
	require.NoError(t, err)
	s.StartProcessMaintenance(ctx)
}
