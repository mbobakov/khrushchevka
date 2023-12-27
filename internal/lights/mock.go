package lights

// TestController is a fake implementation for development without real board
// TestController always returns no error for the set command
type TestController struct {
	IsONFunc func(board uint8, pin string) (bool, error)
	SetFunc  func(board uint8, pin string, isON bool) error
}

func (c *TestController) Set(board uint8, pin string, isON bool) error {
	return c.SetFunc(board, pin, isON)
}

func (c *TestController) IsOn(board uint8, pin string) (bool, error) {
	return c.IsONFunc(board, pin)
}
