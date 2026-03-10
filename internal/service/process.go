package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rakunlabs/query"
	"github.com/rakunlabs/tummy"
	"github.com/worldline-go/types"
)

type ProcessActionRequest struct {
	Action ProcessAction `json:"action"`
}

type ProcessAction string

const (
	ProcessActionTerminate ProcessAction = "terminate"
)

// CreateProcess creates a new process record with running status,
// registers the cancel function, and returns the process ID.
func (s *Service) CreateProcess(ctx context.Context, info ProcessInfo, cancel context.CancelFunc) (string, error) {
	pid := ulid.Make().String()
	now := tummy.Now()

	process := &Process{
		ID:     pid,
		Status: ProcessStatusRunning,
		Info:   info,
		User:   types.NewNull(UserContext(ctx)),
		CreatedAt: types.Time{
			Time: now,
		},
		UpdatedAt: types.Time{
			Time: now,
		},
	}

	if err := s.store.SaveProcess(ctx, process); err != nil {
		return "", fmt.Errorf("save process: %w", err)
	}

	s.RegisterCancel(pid, cancel)

	return pid, nil
}

// CompleteProcess updates a process to completed status with result info.
func (s *Service) CompleteProcess(ctx context.Context, pid string, rowsAffected int64, duration time.Duration) {
	s.CompleteProcessWithCells(ctx, pid, rowsAffected, duration, nil)
}

// CompleteProcessWithCells updates a process to completed status with result info and cell details.
func (s *Service) CompleteProcessWithCells(ctx context.Context, pid string, rowsAffected int64, duration time.Duration, cells []ProcessCellInfo) {
	s.RemoveCancel(pid)

	process, err := s.GetProcessID(ctx, pid)
	if err != nil {
		slog.Error("complete process: get process", "pid", pid, "error", err)
		return
	}

	// Don't overwrite if already terminated
	if process.Status == ProcessStatusTerminated {
		return
	}

	process.Status = ProcessStatusCompleted
	process.Info.RowsAffected = rowsAffected
	process.Info.Duration = duration.Truncate(time.Microsecond).String()
	if len(cells) > 0 {
		process.Info.Cells = cells
	}

	if err := s.store.SaveProcess(ctx, process); err != nil {
		slog.Error("complete process: save process", "pid", pid, "error", err)
	}
}

// FailProcess updates a process to failed status with error info.
func (s *Service) FailProcess(ctx context.Context, pid string, processErr error, duration time.Duration) {
	s.FailProcessWithCells(ctx, pid, processErr, duration, nil)
}

// FailProcessWithCells updates a process to failed status with error info and cell details.
func (s *Service) FailProcessWithCells(ctx context.Context, pid string, processErr error, duration time.Duration, cells []ProcessCellInfo) {
	s.RemoveCancel(pid)

	process, err := s.GetProcessID(ctx, pid)
	if err != nil {
		slog.Error("fail process: get process", "pid", pid, "error", err)
		return
	}

	// Don't overwrite if already terminated
	if process.Status == ProcessStatusTerminated {
		return
	}

	process.Status = ProcessStatusFailed
	process.Info.Error = processErr.Error()
	process.Info.Duration = duration.Truncate(time.Microsecond).String()
	if len(cells) > 0 {
		process.Info.Cells = cells
	}

	if err := s.store.SaveProcess(ctx, process); err != nil {
		slog.Error("fail process: save process", "pid", pid, "error", err)
	}
}

func (s *Service) GetProcess(ctx context.Context, q *query.Query) ([]Process, error) {
	return s.store.GetProcess(ctx, q)
}

func (s *Service) GetProcessID(ctx context.Context, pid string) (*Process, error) {
	v, err := s.store.GetProcess(ctx, query.New().AddWhere(query.NewExpressionCmp(query.OperatorEq, "id", pid)))
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, fmt.Errorf("process with ID %s not found; %w", pid, ErrNotExists)
	}

	process := v[0]

	return &process, nil
}

func (s *Service) ActionProcessID(ctx context.Context, pid string, action ProcessActionRequest) error {
	process, err := s.GetProcessID(ctx, pid)
	if err != nil {
		return err
	}

	switch action.Action {
	case ProcessActionTerminate:
		if process.Status == ProcessStatusCompleted || process.Status == ProcessStatusFailed || process.Status == ProcessStatusTerminated {
			return fmt.Errorf("process with ID %s is already in %s status; %w", pid, process.Status, ErrBadRequest)
		}

		// Cancel the running context
		s.CancelProcess(pid)

		process.Status = ProcessStatusTerminated

		return s.store.SaveProcess(ctx, process)
	default:
		return fmt.Errorf("unsupported action %s; %w", action, ErrBadRequest)
	}
}

// DeleteProcess deletes processes matching the query.
func (s *Service) DeleteProcess(ctx context.Context, q *query.Query) error {
	return s.store.DeleteProcess(ctx, q)
}

// CleanupStaleProcesses marks all running processes as failed on startup.
func (s *Service) CleanupStaleProcesses(ctx context.Context) {
	q := query.New().AddWhere(query.NewExpressionCmp(query.OperatorEq, "status", string(ProcessStatusRunning)))

	processes, err := s.store.GetProcess(ctx, q)
	if err != nil {
		slog.Error("cleanup stale processes: get processes", "error", err)
		return
	}

	for i := range processes {
		processes[i].Status = ProcessStatusFailed
		processes[i].Info.Error = "server restarted"

		if err := s.store.SaveProcess(ctx, &processes[i]); err != nil {
			slog.Error("cleanup stale processes: save process", "pid", processes[i].ID, "error", err)
		}
	}

	if len(processes) > 0 {
		slog.Info("cleaned up stale processes", "count", len(processes))
	}
}
