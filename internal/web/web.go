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
	"github.com/r3labs/sse"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type LightsController interface {
	Set(board uint8, pin string, isON bool) error
	IsOn(board uint8, pin string) (bool, error)
	Subscribe(chan<- internal.PinState)
}

// Server deals with all incomming requests and performs calls to the various internal subsystems
// NB: Page generated base on mapping defined in internal/mapping.go
type Server struct {
	indexTmpl *template.Template
	lights    LightsController
	mapping   [][]internal.Light
	sse       *sse.Server
}

func NewServer(l LightsController, mapping [][]internal.Light) (*Server, error) {
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
		indexTmpl: indexTmpl,
		sse:       sseSrv,
		mapping:   mapping,
	}, nil
}

func (s *Server) Listen(ctx context.Context, addr string) error {
	r := chi.NewRouter()

	r.Get("/", s.index)
	r.Post("/lights/set", s.setLigts)
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
