package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/rakunlabs/logi"
	"github.com/worldline-go/saz/internal/render"
)

type Service struct {
	db    Database
	store Storer
}

func New(db Database, store Storer) *Service {
	return &Service{
		db:    db,
		store: store,
	}
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

func (s *Service) RunNote(ctx context.Context, notePath string, values map[string]any) (err error) {
	if notePath == "" {
		return fmt.Errorf("note path is empty; %w", ErrBadRequest)
	}

	note, err := s.store.GetWithPath(ctx, notePath)
	if err != nil {
		return fmt.Errorf("get note by path %s: %w", notePath, err)
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
		if !note.Content.Cells[i].Enabled.V {
			logi.Ctx(ctx).Info("cell is disabled, skipping execution", logCell)
			continue
		}

		if _, ok := dependency[note.Content.Cells[i].Path.V]; !ok {
			note.Content.Cells[i].Result.V = false
		}

		_, err := s.Run(ctxCell, &note.Content.Cells[i], values, dependency)
		if err != nil {
			return fmt.Errorf("%s; %w", note.Content.Cells[i].Description.V, err)
		}
	}

	return nil
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
