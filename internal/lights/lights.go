package lights

import (
	"fmt"

	"github.com/googolgl/go-mcp23017"
	"github.com/mbobakov/khrushchevka/internal"
)

// On turns on/off the light
func (c *Controller) Set(board uint8, pin string, isON bool) (err error) {
	defer func() {
		for _, ch := range c.notifyCh {
			select {
			case ch <- internal.PinState{
				Addr: internal.LightAddress{
					Board: board,
					Pin:   pin,
				},
				IsOn: isON,
			}: // do nothing
			default: // do nothing
			}
		}
	}()

	drv, ok := c.boards[board]
	if !ok {
		return internal.ErrNoBoardConnected
	}

	if isON {
		err = drv.Set(mcp23017.Pins{pin}).HIGH()
		if err != nil {
			return fmt.Errorf("Couldn't setup pin '%s' to High on '%s'", pin, board)
		}
		return nil
	}

	err = drv.Set(mcp23017.Pins{pin}).LOW()
	if err != nil {
		return fmt.Errorf("Couldn't setup pin '%s' to Low on '%s'", pin, board)
	}

	return nil
}

// IsOn returns true when light is on or false in the oposite case
func (c *Controller) IsOn(board uint8, pin string) (bool, error) {
	drv, ok := c.boards[board]
	if !ok {
		return false, internal.ErrNoBoardConnected
	}

	res, err := drv.Get(mcp23017.Pins{pin})
	if err != nil {
		return false, fmt.Errorf("Couldn't get value for the pin '%s' on '%s'", pin, board)
	}

	val, ok := res[pin]
	if !ok {
		return false, fmt.Errorf("undefined value for the pin '%s' on '%s'", pin, board)
	}

	return val > 0, nil
}

// Subscribe returns a channel to subscribe for the light changes
func (c *Controller) Subscribe(ch chan<- internal.PinState) {
	c.notifyCh = append(c.notifyCh, ch)
}
