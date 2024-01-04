package replay

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mbobakov/khrushchevka/internal"
	"github.com/mbobakov/khrushchevka/internal/lights"
	"github.com/mbobakov/khrushchevka/internal/snapshot"
	"github.com/spf13/afero"
)

type Options struct {
	Showtime   time.Duration `long:"showtime" env:"SHOWTIME" default:"1s" description:"step showtime"`
	ReplayFile string        `long:"replay-file" env:"REPLAY_FILE" default:"./snapshot.json" description:"replay file"`
}

const (
	name = "replay"
)

type Replay struct {
	lights lights.ControllerI
	fs     afero.Fs
	opts   Options
	done   chan struct{}

	mu       sync.Mutex
	isActive bool
}

func New(fs afero.Fs, lights lights.ControllerI, opts Options) *Replay {
	return &Replay{
		fs:     fs,
		lights: lights,
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
	// Open the file
	file, err := r.fs.Open(r.opts.ReplayFile)
	if err != nil {
		return fmt.Errorf("couldn't open file '%s': %w", r.opts.ReplayFile, err)
	}
	defer file.Close()

	for {
		// Create a scanner to read the file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if ctx.Err() != nil || r.IsActive() == false {
				return nil
			}

			data := []snapshot.LightDTO{}
			err := json.Unmarshal(scanner.Bytes(), &data)
			if err != nil {
				return fmt.Errorf("couldn't unmarshal data: %w", err)
			}

			for _, d := range data {
				err := r.lights.Set(internal.LightAddress{Board: d.Board, Pin: d.Pin}, d.IsOn)
				if err != nil {
					return fmt.Errorf("couldn't set light '%v': %w", d, err)
				}
			}

			time.Sleep(r.opts.Showtime)

			for _, d := range data {
				err := r.lights.Set(internal.LightAddress{Board: d.Board, Pin: d.Pin}, false)
				if err != nil {
					return fmt.Errorf("couldn't set light '%v': %w", d, err)
				}
			}

		}

		time.Sleep(r.opts.Showtime)
	}
}

func (r *Replay) IsActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.isActive
}
