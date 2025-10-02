package server

import (
	"context"
	"embed"
	"io/fs"
	"net"
	"net/http"

	"github.com/rakunlabs/ada"
	mfolder "github.com/rakunlabs/ada/handler/folder"
	mcors "github.com/rakunlabs/ada/middleware/cors"
	mlog "github.com/rakunlabs/ada/middleware/log"
	mrecover "github.com/rakunlabs/ada/middleware/recover"
	mrequestid "github.com/rakunlabs/ada/middleware/requestid"
	mtelemetry "github.com/rakunlabs/ada/middleware/telemetry"

	"github.com/worldline-go/saz/internal/config"
	"github.com/worldline-go/saz/internal/service"
)

type Server struct {
	config config.Server

	server  *ada.Server
	service *service.Service
}

//go:embed dist/*
var uiFS embed.FS

func New(ctx context.Context, cfg config.Server, svc *service.Service) (*Server, error) {
	privateToken := cfg.PrivateToken

	mux := ada.New()
	mux.Use(
		mrecover.Middleware(),
		mcors.Middleware(),
		mrequestid.Middleware(),
		mlog.Middleware(),
		mtelemetry.Middleware(),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if privateToken != "" {
					token := r.Header.Get("Private-Token")
					if token == privateToken {
						next.ServeHTTP(w, r)
						return
					}

					ada.NewContext(w, r).
						SetStatus(http.StatusForbidden).
						SendJSON(Response{
							Message: "Forbidden Request",
						})

					return
				}

				next.ServeHTTP(w, r)
			})
		},
	)

	s := &Server{
		config:  cfg,
		server:  mux,
		service: svc,
	}

	// ////////////////////////////////////////////

	baseGroup := mux.Group(cfg.BasePath)
	baseGroup.POST("/api/v1/run", baseGroup.Wrap(s.run))
	baseGroup.POST("/api/v1/run/{note}", baseGroup.Wrap(s.runNote))
	baseGroup.GET("/api/v1/run/{note}", baseGroup.Wrap(s.runNote))

	baseGroup.POST("/api/v1/run/{note}/{cell}", baseGroup.Wrap(s.runNoteCell))
	baseGroup.GET("/api/v1/run/{note}/{cell}", baseGroup.Wrap(s.runNoteCell))

	baseGroup.GET("/api/v1/info", baseGroup.Wrap(s.info))
	baseGroup.GET("/api/v1/notes", baseGroup.Wrap(s.getNotes))
	baseGroup.GET("/api/v1/notes/{id}", baseGroup.Wrap(s.getNote))
	baseGroup.PUT("/api/v1/notes/{id}", baseGroup.Wrap(s.putNote))
	baseGroup.DELETE("/api/v1/notes/{id}", baseGroup.Wrap(s.deleteNote))
	baseGroup.POST("/api/v1/render", baseGroup.Wrap(s.render))

	// ////////////////////////////////////////////

	f, err := fs.Sub(uiFS, "dist")
	if err != nil {
		return nil, err
	}

	folderM, err := mfolder.New(&mfolder.Config{
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
	})
	if err != nil {
		return nil, err
	}

	folderM.SetFs(http.FS(f))

	baseGroup.Handle("/*", folderM)

	// ////////////////////////////////////////////

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	return s.server.StartWithContext(ctx, net.JoinHostPort(s.config.Host, s.config.Port))
}
