package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/rakunlabs/tummy"
	"github.com/worldline-go/conn/database"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/types"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

type Postgres struct {
	db   *sql.DB
	goqu *goqu.Database

	tableNotes exp.IdentifierExpression
	tableCron  exp.IdentifierExpression
}

func New(ctx context.Context, cfg *config.StorePostgres) (*Postgres, error) {
	if cfg == nil {
		return nil, errors.New("postgres configuration is nil")
	}

	if cfg.Migrate.DBTable == "" {
		cfg.Migrate.DBTable = "migrations"
	}
	cfg.Migrate.DBTable = cfg.TablePrefix + cfg.Migrate.DBTable

	if cfg.Migrate.Values == nil {
		cfg.Migrate.Values = make(map[string]string)
	}

	cfg.Migrate.Values["table_prefix"] = cfg.TablePrefix

	if err := MigrateDB(ctx, &cfg.Migrate); err != nil {
		return nil, fmt.Errorf("migrate store postgres: %w", err)
	}

	dbConn, err := database.Connect(ctx, "pgx", cfg.DBDatasource)
	if err != nil {
		return nil, fmt.Errorf("connect to store postgres: %w", err)
	}

	slog.Info("connected to store postgres")

	return conn(cfg, dbConn)
}

func conn(cfg *config.StorePostgres, dbConn *sql.DB) (*Postgres, error) {
	dbGoqu := goqu.New("postgres", dbConn)

	return &Postgres{
		db:         dbConn,
		goqu:       dbGoqu,
		tableNotes: goqu.S(cfg.DBSchema).Table(cfg.TablePrefix + "notes"),
		tableCron:  goqu.S(cfg.DBSchema).Table(cfg.TablePrefix + "cron"),
	}, nil
}

func (s *Postgres) Close() {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			slog.Error("close store postgres connection", "error", err)
		}
	}
}

// ////////////////////////////////////////

func (s *Postgres) Get(ctx context.Context, id string) (*service.Note, error) {
	if id == "" {
		return nil, fmt.Errorf("note ID is empty; %w", service.ErrBadRequest)
	}

	note, err := s.get(ctx, goqu.Ex{"id": id})
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return nil, fmt.Errorf("note with ID %s not found; %w", id, service.ErrNotExists)
		}

		return nil, fmt.Errorf("get note by ID %s: %w", id, err)
	}

	return note, nil
}

func (s *Postgres) GetWithPath(ctx context.Context, path string) (*service.Note, error) {
	if path == "" {
		return nil, fmt.Errorf("note path is empty; %w", service.ErrBadRequest)
	}

	note, err := s.get(ctx, goqu.Ex{"path": path})
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			return nil, fmt.Errorf("note with path %s not found; %w", path, service.ErrNotExists)
		}

		return nil, fmt.Errorf("get note by path %s: %w", path, err)
	}

	return note, nil
}

func (s *Postgres) get(ctx context.Context, ex goqu.Ex) (*service.Note, error) {
	var note Note
	isFound, err := s.goqu.From(s.tableNotes).Where(ex).ScanStructContext(ctx, &note)
	if err != nil {
		return nil, err
	}

	if !isFound {
		return nil, service.ErrNotExists
	}

	return &service.Note{
		ID:        note.ID,
		Name:      note.Name,
		Path:      note.Path,
		Content:   note.Content.V,
		UpdatedBy: note.UpdatedBy,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}, nil
}

func (s *Postgres) Save(ctx context.Context, note *service.Note) error {
	dbNote := Note{
		ID:        note.ID,
		Name:      note.Name,
		Content:   types.NewJSON(note.Content),
		Path:      note.Path,
		UpdatedBy: types.NewNull(service.UserContext(ctx)),
		UpdatedAt: types.NewTimeNull(tummy.Now()),
		CreatedAt: types.NewTimeNull(tummy.Now()),
	}

	// insert or update the note with goqu
	_, err := s.goqu.Insert(s.tableNotes).Rows(dbNote).OnConflict(goqu.DoUpdate("id", dbNote)).Executor().ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("exec upsert note: %w", err)
	}

	return nil
}

func (s *Postgres) GetNotes(ctx context.Context) ([]service.IDName, error) {
	var notes []NoteIDName
	if err := s.goqu.From(s.tableNotes).Select("id", "name").ScanStructsContext(ctx, &notes); err != nil {
		return nil, fmt.Errorf("get notes: %w", err)
	}

	idNames := make([]service.IDName, len(notes))
	for i, note := range notes {
		idNames[i] = service.IDName{
			ID:   note.ID,
			Name: note.Name,
		}
	}

	return idNames, nil
}

func (s *Postgres) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("note ID is empty; %w", service.ErrBadRequest)
	}

	result, err := s.goqu.Delete(s.tableNotes).Where(goqu.Ex{"id": id}).Executor().ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("delete note by ID %s: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected after delete note by ID %s: %w", id, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with ID %s not found; %w", id, service.ErrNotExists)
	}

	return nil
}
