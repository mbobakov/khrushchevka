package lights

import (
	"fmt"

	"github.com/googolgl/go-i2c"
	"github.com/googolgl/go-mcp23017"
	"github.com/mbobakov/khrushchevka/internal"
)

// Callback is the callback function for the lights
// It will be called when the light state changes by Set Command
type Callback func(board uint8, pin string, isOn bool, err error)

// Controller is the controller for the lights
type Controller struct {
	boards   map[uint8]*mcp23017.MCP23017
	notifyCh []chan<- internal.PinState
}

// NewController returns a new controller on the  i2c bus
func NewController(i2cBus string, boards []uint8, callbacks ...Callback) (*Controller, error) {
	if len(i2cBus) == 0 {
		return nil, fmt.Errorf("i2cBus is empty. '/dev/i2c-0' could be a good start")
	}

	cntrl := &Controller{
		boards: make(map[uint8]*mcp23017.MCP23017),
	}

	for _, addr := range boards {
		mcp, err := openMCP23017(addr, i2cBus)
		if err != nil {
			return nil, fmt.Errorf("could open mcp23017 for board %d: %w", addr, err)
		}

		cntrl.boards[addr] = mcp
	}

	return cntrl, nil
}

func openMCP23017(addr uint8, i2cBus string) (*mcp23017.MCP23017, error) {
	i2c, err := i2c.New(addr, i2cBus)
	if err != nil {
		return nil, fmt.Errorf("could not open i2c bus with addr '0x%x': %w", addr, err)
	}

	mcp, err := mcp23017.New(i2c)
	if err != nil {
		return nil, fmt.Errorf("could not create mcp23017 object with addr '0x%x': %w", addr, err)
	}

	err = mcp.Set(mcp23017.AllPins()).OUTPUT()
	if err != nil {
		return nil, fmt.Errorf("could not set all pins to output with addr '0x%x': %w", addr, err)
	}

	err = mcp.Set(mcp23017.AllPins()).LOW()
	if err != nil {
		return nil, fmt.Errorf("could not set all pins to low with addr '0x%x': %w", addr, err)
	}

	return mcp, nil
}
