package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/worldline-go/conn/database"
	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"

	_ "github.com/worldline-go/conn/database/postgres"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

type Postgres struct {
	db   *sqlx.DB
	goqu *goqu.Database
}

func New(ctx context.Context, cfg *config.StorePostgres) (*Postgres, error) {
	if cfg == nil {
		return nil, errors.New("postgres configuration is nil")
	}

	dbConn, err := database.Connect(ctx, "pgx", cfg.Database.DBDatasource)
	if err != nil {
		return nil, fmt.Errorf("connect to store postgres: %w", err)
	}

	dbGoqu := goqu.New("postgres", dbConn)

	slog.Info("connected to store postgres")

	return &Postgres{
		db:   dbConn,
		goqu: dbGoqu,
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
	// var note config.Note
	// query := "SELECT id FROM notes WHERE id = $1"
	// if err := s.db.GetContext(ctx, &note, query, id); err != nil {
	// 	return config.Note{}, fmt.Errorf("get note by id %s: %w", id, err)
	// }
	return note, nil
}
