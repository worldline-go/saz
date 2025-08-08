package config

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rakunlabs/chu"
	"github.com/rakunlabs/logi"
	"github.com/worldline-go/tell"

	_ "github.com/rakunlabs/chu/loader/loaderconsul"
	_ "github.com/rakunlabs/chu/loader/loadervault"
)

var (
	ServerName    = "saz"
	ServerVersion = "v0.0.0"
)

type Config struct {
	LogLevel string              `cfg:"log_level" default:"info"`
	Server   Server              `cfg:"server"`
	Database map[string]Database `cfg:"database"`
	Store    Store               `cfg:"store"`

	Telemetry tell.Config `cfg:"telemetry"`
}

type Server struct {
	BasePath string `cfg:"base_path"`

	Port string `cfg:"port" default:"8080"`
	Host string `cfg:"host"`
}

type Database struct {
	DBDatasource string `cfg:"db_datasource" log:"-"`
	DBType       string `cfg:"db_type"`
	DBSchema     string `cfg:"db_schema"`
}

type Store struct {
	Postgres *StorePostgres `cfg:"postgres"`
}

type Migrate struct {
	DBDatasource string            `cfg:"db_datasource" log:"-"`
	DBType       string            `cfg:"db_type"`
	DBSchema     string            `cfg:"db_schema"`
	DBTable      string            `cfg:"db_table"`
	Values       map[string]string `cfg:"values"`
}

type StorePostgres struct {
	TablePrefix  string  `cfg:"table_prefix"`
	DBDatasource string  `cfg:"db_datasource" log:"-"`
	DBSchema     string  `cfg:"db_schema"`
	Migrate      Migrate `cfg:"migrate"`
}

func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{}
	if err := chu.Load(ctx, "saz", cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := logi.SetLogLevel(cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("set log level %s: %w", cfg.LogLevel, err)
	}

	slog.Info("loaded configuration", "config", chu.MarshalMap(cfg))

	cfg.Telemetry.Logger = slog.Default()

	return cfg, nil
}
