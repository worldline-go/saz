package main

import (
	"context"
	"fmt"

	"github.com/rakunlabs/into"
	"github.com/rakunlabs/logi"

	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/database"
	"github.com/worldline-go/saz/internal/server"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/saz/internal/store"
	"github.com/worldline-go/tell"
)

var (
	version = "v0.0.0"
	commit  = "-"
	date    = "-"
)

func main() {
	config.ServerVersion = version
	into.Init(run,
		into.WithLogger(logi.InitializeLog()),
		into.WithMsgf("%s version:[%s] commit:[%s] date:[%s]", config.ServerName, version, commit, date),
	)
}

func run(ctx context.Context) error {
	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	collector, err := tell.New(ctx, cfg.Telemetry)
	if err != nil {
		return fmt.Errorf("init telemetry; %w", err)
	}

	defer collector.Shutdown()

	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("init database; %w", err)
	}
	defer db.Close()

	st, err := store.New(ctx, cfg.Store)
	if err != nil {
		return fmt.Errorf("init store; %w", err)
	}
	defer st.Close()

	svc := service.New(db, st)

	srv, err := server.New(ctx, cfg.Server, svc)
	if err != nil {
		return fmt.Errorf("init server; %w", err)
	}

	return srv.Start(ctx)
}
