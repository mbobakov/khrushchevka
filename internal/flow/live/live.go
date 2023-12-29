package live

import "context"

type Live string

func (l Live) Name() string {
	return string(l)
}

func (l Live) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (l Live) Stop() {
}
