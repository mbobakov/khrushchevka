package manual

import "context"

type Manual string

func (m Manual) Name() string {
	return string(m)
}

func (m Manual) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (m Manual) Stop() {
}
