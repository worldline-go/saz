package service

import (
	"context"
	"fmt"

	"github.com/rakunlabs/query"
)

type ProcessActionRequest struct {
	Action ProcessAction `json:"action"`
}

type ProcessAction string

const (
	ProcessActionTerminate ProcessAction = "terminate"
)

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

	return &v[0], nil
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

		process.Status = ProcessStatusTerminated
		return s.store.SaveProcess(ctx, process)
	default:
		return fmt.Errorf("unsupported action %s; %w", action, ErrBadRequest)
	}
}
