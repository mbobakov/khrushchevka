package lights

import (
	"fmt"

	"github.com/googolgl/go-mcp23017"
	"github.com/mbobakov/khrushchevka/internal"
)

// On turns on/off the light for the provided side and level and number
func (c *Controller) Set(board uint8, pin string, isON bool) error {
	drv, ok := c.boards[board]
	if !ok {
		return internal.ErrNoBoardConnected
	}

	if isON {
		err := drv.Set(mcp23017.Pins{pin}).HIGH()
		if err != nil {
			return fmt.Errorf("Couldn't setup pin '%s' to High on '%s'", board, pin)
		}
	}

	err := drv.Set(mcp23017.Pins{pin}).LOW()
	if err != nil {
		return fmt.Errorf("Couldn't setup pin '%s' to Low on '%s'", board, pin)
	}

	return nil
}
