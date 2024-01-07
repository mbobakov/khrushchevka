package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/mbobakov/khrushchevka/internal"
	"github.com/mbobakov/khrushchevka/internal/lights"
	"github.com/r3labs/sse"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type FlowController interface {
	SelectFlow(ctx context.Context, name string) error
	FlowNames() []string
	Active() string
}

type Snapshoter interface {
	Snapshot() error
}

// Server deals with all incomming requests and performs calls to the various internal subsystems
// NB: Page generated base on mapping defined in internal/mapping.go
type Server struct {
	indexTmpl           *template.Template
	lights              lights.ControllerI
	flows               FlowController
	snap                Snapshoter
	mapping             [][]internal.Light
	sse                 *sse.Server
	mainCtx             context.Context
	validateSelectBoard uint8
}

func NewServer(l lights.ControllerI, f FlowController, snap Snapshoter, mapping [][]internal.Light) (*Server, error) {
	// templates
	indexTmpl, err := template.ParseFS(templatesFS, "templates/*.gotmpl")
	if err != nil {
		return nil, fmt.Errorf("couldn't parse index templates: %w", err)
	}

	sseSrv := sse.New()
	sseSrv.AutoReplay = false
	sseSrv.BufferSize = 0
	sseSrv.EventTTL = 0
	sseSrv.CreateStream("lights")

	return &Server{
		lights:    l,
		flows:     f,
		indexTmpl: indexTmpl,
		sse:       sseSrv,
		mapping:   mapping,
		snap:      snap,
	}, nil
}

func (s *Server) Listen(ctx context.Context, addr string) error {
	s.mainCtx = ctx

	r := chi.NewRouter()

	r.Get("/", s.index)

	r.Get("/validate", s.validate)
	r.Post("/validate", s.validatePost)

	r.Post("/lights/set", s.setLigts)
	r.Post("/lights/snapshot", s.snapshot)

	r.Put("/flows", s.setFlow)

	r.Get("/static/*", http.FileServer(http.FS(staticFS)).ServeHTTP)

	r.Get("/events", s.sse.HTTPHandler)

	server := http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: time.Second,
	}

	go stopWhenDone(ctx, &server)

	return server.ListenAndServe()
}

func stopWhenDone(ctx context.Context, server *http.Server) {
	<-ctx.Done()
	err := server.Shutdown(context.Background())
	if err != nil {
		slog.Error("Failed to gracefully shutdown HTTP server:%v", err)
	}
}

func lightID(l internal.Light) string {
	return fmt.Sprintf("l-%d-%s", l.Addr.Board, l.Addr.Pin)
}
