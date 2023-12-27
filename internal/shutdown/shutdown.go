package shutdown

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// ErrGracefulShutdown indicates a graceful shutdown happened
var ErrGracefulShutdown = errors.New("graceful shutdown")

// Receive will block waiting for an interupt or termination signal and return a graceful shutdown error
func Receive(ctx context.Context) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sigs:
			// nolint: wrapcheck
			return ErrGracefulShutdown
		}
	}
}
