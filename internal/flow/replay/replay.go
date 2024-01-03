package replay

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Options struct {
	Showtime time.Duration `long:"showtime" env:"SHOWTIME" default:"1s" description:"step showtime"`
}

type replayer interface {
	Replay(ctx context.Context, filpath string, delay time.Duration) error
}

const (
	name = "replay"
)

type Replay struct {
	replay replayer
	opts   Options
	done   chan struct{}

	mu       sync.Mutex
	isActive bool
}

func New(r replayer, opts Options) *Replay {
	return &Replay{
		replay: r,
		opts:   opts,
		done:   make(chan struct{}),
	}
}

func (r *Replay) Name() string {
	return name
}

func (r *Replay) Start(ctx context.Context) error {
	slog.Info("starting flow", "flow", name)
	r.done = make(chan struct{})
	r.mu.Lock()
	r.isActive = true
	r.mu.Unlock()

	return r.mainCycle(ctx)
}

func (m *Replay) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isActive {
		slog.Info("stopping flow", "flow", name)
		m.isActive = false
		close(m.done)
	}
}

func (r *Replay) mainCycle(ctx context.Context) error {
	rerr := make(chan error)
	go func() {
		rerr <- r.replay.Replay(ctx, "./snapshot.json", r.opts.Showtime)
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.done:
			return nil
		case err := <-rerr:
			if err != nil {
				return fmt.Errorf("couldn't replay: %w", err)
			}
			go func() {
				rerr <- r.replay.Replay(ctx, "./snapshot.json", r.opts.Showtime)
			}()
		}
	}
}
