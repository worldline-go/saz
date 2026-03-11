package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/rakunlabs/alan"
	"github.com/rakunlabs/logi"
	"github.com/worldline-go/saz/internal/render"
)

type Service struct {
	db    Database
	store Storer

	cancelMu  sync.Mutex
	cancelMap map[string]context.CancelFunc

	alan *alan.Alan
}

// peerMessage is the message format exchanged between instances via alan.
type peerMessage struct {
	Action string `json:"action"`
	PID    string `json:"pid"`
}

// peerResponse is the response format for alan peer messages.
type peerResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func New(db Database, store Storer, cfg *alan.Config) (*Service, error) {
	s := &Service{
		db:        db,
		store:     store,
		cancelMap: make(map[string]context.CancelFunc),
	}

	if cfg != nil {
		a, err := alan.New(*cfg)
		if err != nil {
			return nil, fmt.Errorf("init alan: %w", err)
		}

		s.alan = a
		slog.Info("alan distributed communication enabled", "dns_addr", cfg.DNSAddr, "port", cfg.Port)
	}

	return s, nil
}

// StartAlan starts the alan peer communication. Blocks until ctx is cancelled.
// Returns nil immediately if alan is not configured.
func (s *Service) StartAlan(ctx context.Context) error {
	if s.alan == nil {
		return nil
	}

	return s.alan.Start(ctx, s.handlePeerMessage)
}

// StopAlan gracefully stops alan peer communication.
func (s *Service) StopAlan() error {
	if s.alan == nil {
		return nil
	}

	return s.alan.Stop()
}

// handlePeerMessage processes incoming messages from alan peers.
func (s *Service) handlePeerMessage(_ context.Context, msg alan.Message) {
	if !msg.IsRequest() {
		return
	}

	var req peerMessage
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		slog.Error("alan: failed to unmarshal peer message", "error", err)
		resp, _ := json.Marshal(peerResponse{OK: false, Error: "invalid message"})
		s.alan.Reply(msg, resp)
		return
	}

	switch req.Action {
	case "terminate":
		found := s.cancelProcessLocal(req.PID)
		resp, _ := json.Marshal(peerResponse{OK: found})
		s.alan.Reply(msg, resp)

		if found {
			slog.Info("alan: terminated process from peer request", "pid", req.PID)
		}
	default:
		resp, _ := json.Marshal(peerResponse{OK: false, Error: "unknown action"})
		s.alan.Reply(msg, resp)
	}
}

// RegisterCancel stores a cancel function for a process ID.
func (s *Service) RegisterCancel(pid string, cancel context.CancelFunc) {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	s.cancelMap[pid] = cancel
}

// CancelProcess invokes and removes the cancel function for a process ID.
// If not found locally and alan is configured, broadcasts to peers.
// Returns true if a cancel function was found and invoked (locally or on a peer).
func (s *Service) CancelProcess(pid string) bool {
	// Fast path: try local first
	if s.cancelProcessLocal(pid) {
		return true
	}

	// Distributed path: ask peers via alan
	if s.alan == nil {
		return false
	}

	req, _ := json.Marshal(peerMessage{Action: "terminate", PID: pid})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	replies, err := s.alan.SendAndWaitReply(ctx, req)
	if err != nil {
		slog.Error("alan: failed to broadcast terminate", "pid", pid, "error", err)
		return false
	}

	for _, reply := range replies {
		var resp peerResponse
		if err := json.Unmarshal(reply.Data, &resp); err != nil {
			continue
		}
		if resp.OK {
			slog.Info("alan: process terminated by peer", "pid", pid, "peer", reply.Addr)
			return true
		}
	}

	return false
}

// cancelProcessLocal invokes and removes the cancel function from the local map only.
func (s *Service) cancelProcessLocal(pid string) bool {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	if cancel, ok := s.cancelMap[pid]; ok {
		cancel()
		delete(s.cancelMap, pid)
		return true
	}
	return false
}

// RemoveCancel removes the cancel function for a process ID without invoking it.
func (s *Service) RemoveCancel(pid string) {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	delete(s.cancelMap, pid)
}

func (s *Service) Run(ctx context.Context, cell *Cell, values map[string]any, dependency map[string]struct{}) (result Result, err error) {
	if cell == nil || cell.DBType == "" || cell.Content == "" {
		return nil, fmt.Errorf("invalid cell; %w", ErrBadRequest)
	}

	logCell := slog.Group("cell",
		slog.String("description", cell.Description.V),
		slog.String("db_type", cell.DBType),
		slog.String("mode", cell.Mode.V.Name),
	)
	logi.Ctx(ctx).Info("running cell", logCell)

	defer func() {
		if err != nil {
			logi.Ctx(ctx).Error("failed to run cell", logCell, slog.String("error", err.Error()))
		} else {
			logi.Ctx(ctx).Info("cell executed successfully",
				logCell,
				slog.Int64("row_affected", result.RowsAffected()),
				slog.String("duration", result.Duration().String()),
			)
		}
	}()

	content := cell.Content
	if cell.Template.Enabled {
		contentRendered, err := render.ExecuteWithData(content, values)
		if err != nil {
			return nil, fmt.Errorf("render content: %w", err)
		}

		content = string(contentRendered)
	}

	if cell.Mode.V.Enabled {
		switch cell.Mode.V.Name {
		case "transfer":
			if cell.Mode.V.Table == "" {
				return nil, fmt.Errorf("transfer mode requires a table name; %w", ErrBadRequest)
			}
			columns, iterGet, err := s.db.IterGet(ctx, cell.DBType, content, cell.Mode.V.MapType)
			if err != nil {
				return nil, fmt.Errorf("get iterator: %w", err)
			}

			// TODO: make better handling of iterators
			defer func() {
				for range iterGet {
					return
				}
			}()

			result, err := s.db.IterSet(ctx, cell.Mode.V.DBType, cell.Mode.V.Table, cell.Mode.V.Wipe, cell.Mode.V.SkipError, cell.Mode.V.MapType, cell.Mode.V.Batch, columns, iterGet)
			if err != nil {
				return nil, fmt.Errorf("set iterator: %w", err)
			}

			return result, nil
		default:
			return nil, fmt.Errorf("unsupported mode %s; %w", cell.Mode.V.Name, ErrBadRequest)
		}
	}

	if cell.Result.V {
		result, err := s.db.Query(ctx, cell.DBType, content, cell.Limit)
		if err != nil {
			return nil, err
		}

		if cell.Path.V != "" && len(dependency) > 0 {
			if _, ok := dependency[cell.Path.V]; ok {
				if cells, exists := values["cells"]; exists {
					cellsMap, ok := cells.(map[string]any)
					if !ok {
						return nil, fmt.Errorf("invalid cells dependency format; %w", ErrBadRequest)
					}
					cellsMap[cell.Path.V] = DataToMap(result.Columns(), result.Rows())
					values["cells"] = cellsMap
				} else {
					values["cells"] = map[string]any{
						cell.Path.V: DataToMap(result.Columns(), result.Rows()),
					}
				}
			}
		}

		return result, nil
	}

	return s.db.Exec(ctx, cell.DBType, content)
}

func (s *Service) RunNote(ctx context.Context, notePath string, values map[string]any) (cellInfos []ProcessCellInfo, err error) {
	if notePath == "" {
		return nil, fmt.Errorf("note path is empty; %w", ErrBadRequest)
	}

	note, err := s.store.GetWithPath(ctx, notePath)
	if err != nil {
		return nil, fmt.Errorf("get note by path %s: %w", notePath, err)
	}

	// get all dependencies
	dependency := make(map[string]struct{})
	for i := range note.Content.Cells {
		if note.Content.Cells[i].Dependency.V.Enabled {
			for _, name := range note.Content.Cells[i].Dependency.V.Names {
				dependency[name] = struct{}{}
			}
		}
	}

	logNote := slog.Group("note", slog.String("name", note.Name), slog.String("path", note.Path))

	defer func() {
		if err != nil {
			logi.Ctx(ctx).Error("failed to run note", logNote, slog.String("error", err.Error()))
		} else {
			logi.Ctx(ctx).Info("note executed successfully", logNote)
		}
	}()

	logi.Ctx(ctx).Info("starting note execution", logNote)

	for i := range note.Content.Cells {
		logCell := slog.Group("cell", slog.String("description", note.Content.Cells[i].Description.V), slog.Int("number", i+1))
		ctxCell := logi.WithContext(ctx, logi.Ctx(ctx).With(logNote, logCell))

		cellInfo := ProcessCellInfo{
			Description: note.Content.Cells[i].Description.V,
			Query:       note.Content.Cells[i].Content,
		}

		if !note.Content.Cells[i].Enabled.V {
			logi.Ctx(ctx).Info("cell is disabled, skipping execution", logCell)
			cellInfo.Status = "skipped"
			cellInfos = append(cellInfos, cellInfo)
			continue
		}

		if _, ok := dependency[note.Content.Cells[i].Path.V]; !ok {
			note.Content.Cells[i].Result.V = false
		}

		cellStart := time.Now()
		result, runErr := s.Run(ctxCell, &note.Content.Cells[i], values, dependency)
		cellInfo.Duration = time.Since(cellStart).Truncate(time.Microsecond).String()

		if runErr != nil {
			cellInfo.Status = "failed"
			cellInfo.Error = runErr.Error()
			cellInfos = append(cellInfos, cellInfo)
			return cellInfos, fmt.Errorf("%s; %w", note.Content.Cells[i].Description.V, runErr)
		}

		cellInfo.Status = "completed"
		cellInfo.RowsAffected = result.RowsAffected()
		cellInfos = append(cellInfos, cellInfo)
	}

	return cellInfos, nil
}

func (s *Service) RunNoteCell(ctx context.Context, notePath string, cellPath string, values map[string]any) (result Result, err error) {
	if notePath == "" {
		return nil, fmt.Errorf("note path is empty; %w", ErrBadRequest)
	}
	if cellPath == "" {
		return nil, fmt.Errorf("cell is invalid; %w", ErrBadRequest)
	}

	note, err := s.store.GetWithPath(ctx, notePath)
	if err != nil {
		return nil, fmt.Errorf("get note by path %s: %w", notePath, err)
	}

	var cellNode *Cell
	for i := range note.Content.Cells {
		if note.Content.Cells[i].Path.V == cellPath {
			cellNode = &note.Content.Cells[i]
			break
		}
	}

	if cellNode == nil {
		cellNumber, err := strconv.Atoi(cellPath)
		if err != nil || cellNumber < 1 {
			return nil, fmt.Errorf("invalid cell number; %w", ErrBadRequest)
		}

		cellNode = &note.Content.Cells[cellNumber-1]
	}

	if cellNode == nil {
		return nil, fmt.Errorf("cell %s not found in note %s; %w", cellPath, notePath, ErrNotExists)
	}

	logNote := slog.Group("note", slog.String("name", note.Name), slog.String("path", note.Path))
	logCell := slog.Group("cell", slog.String("description", cellNode.Description.V), slog.String("path", cellPath))
	ctxCell := logi.WithContext(ctx, logi.Ctx(ctx).With(logNote, logCell))

	defer func() {
		if err != nil {
			logi.Ctx(ctxCell).Error("failed to run cell", logNote, logCell, slog.String("error", err.Error()))
		} else {
			logi.Ctx(ctxCell).Info("cell executed successfully", logNote, logCell)
		}
	}()

	logi.Ctx(ctxCell).Info("starting cell execution", logNote, logCell)

	dependency := make(map[string]struct{})
	if cellNode.Dependency.V.Enabled {
		for _, name := range cellNode.Dependency.V.Names {
			dependency[name] = struct{}{}
		}
	}

	for name := range dependency {
		var depCell *Cell
		for i := range note.Content.Cells {
			if note.Content.Cells[i].Path.V == name {
				depCell = &note.Content.Cells[i]
				break
			}
		}

		if depCell == nil {
			return nil, fmt.Errorf("dependency cell %s not found in note %s; %w", name, notePath, ErrNotExists)
		}

		if _, err := s.Run(ctxCell, depCell, values, dependency); err != nil {
			return nil, fmt.Errorf("execute dependency cell %s: %w", name, err)
		}
	}

	result, err = s.Run(ctxCell, cellNode, values, nil)
	if err != nil {
		return nil, err
	}

	if cellNode.Result.V {
		logi.Ctx(ctxCell).Info("cell result",
			slog.Int64("rows_affected", result.RowsAffected()),
			slog.Int("columns", len(result.Columns())),
			slog.Int("rows", len(result.Rows())),
			slog.String("duration", result.Duration().String()),
		)
	}

	return result, nil
}

func (s *Service) DatabaseList() []string {
	return s.db.DatabaseList()
}

func (s *Service) GetNote(ctx context.Context, id string) (*Note, error) {
	return s.store.Get(ctx, id)
}

func (s *Service) SaveNote(ctx context.Context, note *Note) error {
	if note == nil {
		return fmt.Errorf("note is nil; %w", ErrBadRequest)
	}
	if note.ID == "" || note.Name == "" || note.Path == "" {
		return fmt.Errorf("invalid Name, ID, or Path; %w", ErrBadRequest)
	}

	return s.store.Save(ctx, note)
}

func (s *Service) DeleteNote(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("note ID is empty; %w", ErrBadRequest)
	}

	return s.store.Delete(ctx, id)
}

func (s *Service) GetNotes(ctx context.Context) ([]IDName, error) {
	return s.store.GetNotes(ctx)
}
