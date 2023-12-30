package lights

import (
	"fmt"
	"log/slog"

	"github.com/googolgl/go-mcp23017"
	"github.com/mbobakov/khrushchevka/internal"
)

// On turns on/off the light
func (c *Controller) Set(addr internal.LightAddress, isON bool) (err error) {
	l := slog.With("controller", "mcp23017")
	defer func() {
		for _, ch := range c.notifyCh {
			select {
			case ch <- internal.PinState{
				Addr: internal.LightAddress{
					Board: addr.Board,
					Pin:   addr.Pin,
				},
				IsOn: isON,
			}: // do nothing
			default: // do nothing
			}
		}
	}()

	drv, ok := c.boards[addr.Board]
	if !ok {
		return internal.ErrNoBoardConnected
	}
	l.Info("setting light", slog.Int("board", int(addr.Board)), slog.String("pin", addr.Pin), slog.Bool("isON", isON))
	if isON {
		err = drv.Set(mcp23017.Pins{addr.Pin}).HIGH()
		if err != nil {
			return fmt.Errorf("Couldn't setup pin '%s' to High on '%d'", addr.Pin, addr.Board)
		}
		return nil
	}

	err = drv.Set(mcp23017.Pins{addr.Pin}).LOW()
	if err != nil {
		return fmt.Errorf("Couldn't setup pin '%s' to Low on '%d'", addr.Pin, addr.Board)
	}

	return nil
}

// IsOn returns true when light is on or false in the oposite case
func (c *Controller) IsOn(addr internal.LightAddress) (bool, error) {
	drv, ok := c.boards[addr.Board]
	if !ok {
		return false, internal.ErrNoBoardConnected
	}

	res, err := drv.Get(mcp23017.Pins{addr.Pin})
	if err != nil {
		return false, fmt.Errorf("Couldn't get value for the pin '%s' on '%d'", addr.Pin, addr.Board)
	}

	val, ok := res[addr.Pin]
	if !ok {
		return false, fmt.Errorf("undefined value for the pin '%s' on '%d'", addr.Pin, addr.Board)
	}

	return val > 0, nil
}

// Subscribe returns a channel to subscribe for the light changes
func (c *Controller) Subscribe(ch chan<- internal.PinState) {
	c.notifyCh = append(c.notifyCh, ch)
}
