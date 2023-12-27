package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"

	"github.com/jessevdk/go-flags"
	"github.com/mbobakov/khrushchevka/internal/lights"
	"github.com/mbobakov/khrushchevka/internal/shutdown"
	"github.com/mbobakov/khrushchevka/internal/web"
	"golang.org/x/sync/errgroup"
)

type options struct {
	Listen string  `long:"listen" env:"LISTEN" default:":8080" description:"Listen address"`
	Boards []uint8 `long:"boards" env:"BOARDS" default:"20,21,22,23,24,25" env-delim:"," description:"Boards to validate"`
	NoOp   bool    `long:"noop" env:"NOOP" description:"If true fake board will be used"`
}

func main() {
	opts := options{}
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return
		}
		log.Fatalf("Cannot parse flags :%v", err)
	}
	appctx := context.Background()
	err := realMain(appctx, opts)
	if err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func realMain(appctx context.Context, opts options) error {
	var (
		prov web.LightsController
		err  error
	)

	prov = &lights.TestController{
		IsONFunc: func(board uint8, pin string) (bool, error) { return rand.Intn(2) == 1, nil },
		SetFunc:  func(board uint8, pin string, isON bool) error { return nil },
	}

	if !opts.NoOp {
		prov, err = lights.NewController("/dev/i2c-0", opts.Boards)
		if err != nil {
			return fmt.Errorf("couldn't initiate controller for the boards: %w", err)
		}
	}

	g, ctx := errgroup.WithContext(appctx)
	srv, err := web.NewServer(prov)
	if err != nil {
		return fmt.Errorf("couln't initiate web server: %w", err)
	}

	g.Go(func() error { return srv.Listen(ctx, opts.Listen) })

	g.Go(func() error { return shutdown.Receive(ctx) })

	slog.Info("Service started and listens on '%s'", opts.Listen)
	err = g.Wait()

	if err != nil && errors.Is(err, shutdown.ErrGracefulShutdown) {
		slog.Info("Service shutted down successfuly")
		return nil
	}

	slog.Error("Service shut down unexpectedly: %v", err)

	return err
}
