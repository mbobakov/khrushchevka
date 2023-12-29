package lights

import "github.com/mbobakov/khrushchevka/internal"

// TestController is a fake implementation for development without real board
// TestController always returns no error for the set command
type TestController struct {
	IsONFunc func(board uint8, pin string) (bool, error)
	SetFunc  func(board uint8, pin string, isON bool) error
	NotifyCh []chan<- internal.PinState
}

func (c *TestController) Set(board uint8, pin string, isON bool) error {
	defer func() {
		for _, ch := range c.NotifyCh {
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

	return c.SetFunc(board, pin, isON)
}

func (c *TestController) IsOn(board uint8, pin string) (bool, error) {
	return c.IsONFunc(board, pin)
}

func (c *TestController) Subscribe(ch chan<- internal.PinState) {
	c.NotifyCh = append(c.NotifyCh, ch)
}
