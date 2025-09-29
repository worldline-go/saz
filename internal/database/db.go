package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/worldline-go/conn/database"
	"github.com/worldline-go/saz/internal/config"

	_ "github.com/alexbrainman/odbc"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/godror/godror"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/worldline-go/conn/database/postgres"
)

type Database struct {
	DB map[string]*Info
}

type Info struct {
	DB          *sqlx.DB
	PlaceHolder string
}

func (d *Database) Close() {
	for name, dbConn := range d.DB {
		if err := dbConn.DB.Close(); err != nil {
			slog.Error("close database connection", "name", name, "error", err)
		}
	}
}

func Connect(ctx context.Context, cfg map[string]config.Database) (*Database, error) {
	db := &Database{
		DB: make(map[string]*Info),
	}

	for name, dbConfig := range cfg {
		dbConn, err := database.Connect(ctx, dbConfig.DBType, dbConfig.DBDatasource)
		if err != nil {
			db.Close()

			return nil, fmt.Errorf("connect to database %s: %w", name, err)
		}

		slog.Info("connected to database", "name", name, "type", dbConfig.DBType)

		db.DB[name] = &Info{
			DB:          dbConn,
			PlaceHolder: PlaceHolder(dbConfig.DBType),
		}
	}

	return db, nil
}

func PlaceHolder(dbType string) string {
	switch dbType {
	case "pgx", "postgres":
		return "$"
	case "godror":
		return ":"
	default:
		return "?"
	}
}
