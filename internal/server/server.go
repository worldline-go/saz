package server

import (
	"context"
	"embed"
	"io/fs"
	"net"
	"net/http"

	"github.com/rakunlabs/ada"
	mcors "github.com/rakunlabs/ada/middleware/cors"
	mfolder "github.com/rakunlabs/ada/middleware/folder"
	mlog "github.com/rakunlabs/ada/middleware/log"
	mrecover "github.com/rakunlabs/ada/middleware/recover"
	mrequestid "github.com/rakunlabs/ada/middleware/requestid"

	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/tell/metric/metrichttp"
	"github.com/worldline-go/tell/trace/tracehttp"
)

type Server struct {
	config config.Server

	server  *ada.Server
	service *service.Service
}

//go:embed dist/*
var uiFS embed.FS

func New(ctx context.Context, cfg config.Server, svc *service.Service) (*Server, error) {
	mux := ada.New()
	mux.Use(
		mrecover.Middleware(),
		mcors.Middleware(),
		mrequestid.Middleware(),
		mlog.Middleware(),
		metrichttp.Middleware(),
		tracehttp.Middleware(),
	)

	s := &Server{
		config:  cfg,
		server:  mux,
		service: svc,
	}

	// ////////////////////////////////////////////

	baseGroup := mux.Group(cfg.BasePath)
	baseGroup.POST("/api/v1/run", s.run)
	baseGroup.POST("/api/v1/run/{note}", s.runNote)
	baseGroup.GET("/api/v1/info", s.info)
	baseGroup.GET("/api/v1/notes", s.getNotes)
	baseGroup.GET("/api/v1/notes/{id}", s.getNote)
	baseGroup.PUT("/api/v1/notes/{id}", s.putNote)

	// ////////////////////////////////////////////

	f, err := fs.Sub(uiFS, "dist")
	if err != nil {
		return nil, err
	}

	folderM := mfolder.Folder{
		BasePath:       cfg.BasePath,
		Index:          true,
		StripIndexName: true,
		SPA:            true,
		PrefixPath:     cfg.BasePath,
		CacheRegex: []*mfolder.RegexCacheStore{
			{
				Regex:        `index\.html$`,
				CacheControl: "no-store",
			},
		},
	}

	folderM.SetFs(http.FS(f))
	folderHandler, err := folderM.Handler()
	if err != nil {
		return nil, err
	}

	baseGroup.HandleFunc("/*", folderHandler)

	// ////////////////////////////////////////////

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	return s.server.StartWithContext(ctx, net.JoinHostPort(s.config.Host, s.config.Port))
}
