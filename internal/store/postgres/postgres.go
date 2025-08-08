package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/worldline-go/conn/database"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/types"

	_ "github.com/worldline-go/conn/database/postgres"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

type Postgres struct {
	db     *sqlx.DB
	goqu   *goqu.Database
	prefix string
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

	dbGoqu := goqu.New("postgres", dbConn)

	slog.Info("connected to store postgres")

	return &Postgres{
		db:     dbConn,
		goqu:   dbGoqu,
		prefix: cfg.TablePrefix,
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

	var note Note
	isFound, err := s.goqu.From(s.prefix+"notes").Where(goqu.Ex{"id": id}).ScanStructContext(ctx, &note)
	if err != nil {
		return nil, fmt.Errorf("get note by ID %s: %w", id, err)
	}

	if !isFound {
		return nil, fmt.Errorf("note with ID %s not found; %w", id, service.ErrNotExists)
	}

	return &service.Note{
		ID:      note.ID,
		Name:    note.Name,
		Path:    note.Path,
		Content: note.Content.V,
	}, nil
}

func (s *Postgres) Save(ctx context.Context, note *service.Note) error {
	dbNote := Note{
		ID:        note.ID,
		Name:      note.Name,
		Content:   types.NewJSON(note.Content),
		Path:      note.Path,
		UpdatedBy: types.NewNull(service.UserContext(ctx)),
		UpdatedAt: types.NewTimeNull(time.Now()),
		CreatedAt: types.NewTimeNull(time.Now()),
	}

	// insert or update the note with goqu
	_, err := s.goqu.Insert(s.prefix + "notes").Rows(dbNote).OnConflict(goqu.DoUpdate("id", dbNote)).Executor().ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("exec upsert note: %w", err)
	}

	return nil
}

func (s *Postgres) GetNotes(ctx context.Context) ([]service.IDName, error) {
	var notes []NoteIDName
	if err := s.goqu.From(s.prefix+"notes").Select("id", "name").ScanStructsContext(ctx, &notes); err != nil {
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
