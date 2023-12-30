package lights

import (
	"log/slog"
	"sync"

	"github.com/mbobakov/khrushchevka/internal"
)

// TestController is a fake implementation for development without real board
// TestController always returns no error for the set command
type TestController struct {
	notifyCh []chan<- internal.PinState

	mu    sync.RWMutex
	state map[internal.LightAddress]bool
}

func NewTestController() *TestController {
	return &TestController{
		state: make(map[internal.LightAddress]bool),
	}
}

func (c *TestController) Set(addr internal.LightAddress, isON bool) error {
	l := slog.With("controller", "test")

	defer func() {
		for _, ch := range c.notifyCh {
			ch <- internal.PinState{
				Addr: internal.LightAddress{
					Board: addr.Board,
					Pin:   addr.Pin,
				},
				IsOn: isON,
			}
		}
	}()

	l.Debug("setting light", slog.Int("board", int(addr.Board)), slog.String("pin", addr.Pin), slog.Bool("isON", isON))
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state[addr] = isON

	return nil
}

func (c *TestController) IsOn(addr internal.LightAddress) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state[addr], nil
}

func (c *TestController) Subscribe(ch chan<- internal.PinState) {
	c.notifyCh = append(c.notifyCh, ch)
}
