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

	Telemetry tell.Config `cfg:"telemetry"`
}

type Server struct {
	BasePath string `cfg:"base_path"`

	Port string `cfg:"port" default:"8080"`
	Host string `cfg:"host"`
}

type Database struct {
	DBType       string `cfg:"db_type"`
	DBDatasource string `cfg:"db_datasource"`
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
