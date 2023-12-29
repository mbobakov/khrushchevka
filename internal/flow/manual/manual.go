package manual

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mbobakov/khrushchevka/internal"
)

type LightsController interface {
	Set(addr internal.LightAddress, isON bool) error
	IsOn(addr internal.LightAddress) (bool, error)
}

const (
	name = "manual"
)

type Manual struct {
	lights  LightsController
	mapping [][]internal.Light
	done    chan struct{}

	mu       sync.Mutex
	isActive bool
}

func New(l LightsController, mapping [][]internal.Light) *Manual {
	return &Manual{
		lights:  l,
		mapping: mapping,
		done:    make(chan struct{}),
	}
}

func (m *Manual) Name() string {
	return name
}

func (m *Manual) Start(ctx context.Context) error {
	slog.Info("starting flow", "flow", name)
	m.done = make(chan struct{})
	m.mu.Lock()
	m.isActive = true
	m.mu.Unlock()

	// switch of all lights
	slog.Info("swithching off all lights")
	for _, row := range m.mapping {
		for _, light := range row {
			err := m.lights.Set(light.Addr, false)

			if err != nil {
				return fmt.Errorf("couldn't switch off light '%v': %w", light.Addr, err)
			}
		}
	}

	select {
	case <-ctx.Done():
		return nil
	case <-m.done:
		return nil
	}
}

func (m *Manual) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isActive {
		slog.Info("stopping flow", "flow", name)
		m.isActive = false
		close(m.done)
	}
}
