package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/worldline-go/conn/database"
	"github.com/worldline-go/igmigrator/v2"
	"github.com/worldline-go/saz/internal/config"
)

//go:embed migrations/*
var migrationFS embed.FS

func MigrateDB(ctx context.Context, cfg *config.Migrate) error {
	if cfg.DBDatasource == "" {
		return errors.New("migrate database datasource is empty")
	}

	migration, err := fs.Sub(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate database fs sub: %w", err)
	}

	db, err := database.Connect(ctx, "pgx", cfg.DBDatasource)
	if err != nil {
		return fmt.Errorf("migrate database connect: %w", err)
	}

	defer db.Close()

	result, err := igmigrator.Migrate(ctx, db, &igmigrator.Config{
		Migrations:     migration,
		Schema:         cfg.DBSchema,
		MigrationTable: cfg.DBTable,
		Logger:         slog.Default(),
		Values:         cfg.Values,
	})
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	for mPath, m := range result.Path {
		if m.NewVersion != m.PrevVersion {
			slog.Info("migration applied", "path", mPath, "from_version", m.PrevVersion, "to_version", m.NewVersion)
		}
	}

	return nil
}
