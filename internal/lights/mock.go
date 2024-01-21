package lights

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/mbobakov/khrushchevka/internal"
)

var _ ControllerI = &TestController{}

// TestController is a fake implementation for development without real board
// TestController always returns no error for the set command
type TestController struct {
	notifyCh []chan<- internal.PinState
	boards   []uint8

	mu    sync.RWMutex
	state map[internal.LightAddress]bool
}

func NewTestController(boards []uint8) *TestController {
	for _, b := range boards {
		slog.Info("opened mock", slog.Int("board", int(b)), slog.String("hex", fmt.Sprintf("0x%x", b)))
	}

	return &TestController{
		boards: boards,
		state:  make(map[internal.LightAddress]bool),
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

func (c *TestController) Boards() []uint8 {
	return c.boards
}

func (c *TestController) Reset() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = make(map[internal.LightAddress]bool)
	return nil
}
