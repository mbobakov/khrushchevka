package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/jessevdk/go-flags"
	"github.com/mbobakov/khrushchevka/internal"
	"github.com/mbobakov/khrushchevka/internal/flow"
	"github.com/mbobakov/khrushchevka/internal/flow/live"
	"github.com/mbobakov/khrushchevka/internal/flow/manual"
	"github.com/mbobakov/khrushchevka/internal/flow/replay"
	"github.com/mbobakov/khrushchevka/internal/lights"
	"github.com/mbobakov/khrushchevka/internal/shutdown"
	"github.com/mbobakov/khrushchevka/internal/snapshot/file"
	"github.com/mbobakov/khrushchevka/internal/web"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
)

type options struct {
	Listen string         `long:"listen" env:"LISTEN" default:":8080" description:"Listen address"`
	Boards []uint8        `long:"boards" env:"BOARDS" default:"20,21,22,23,24,25" env-delim:"," description:"Boards to validate"`
	NoOp   bool           `long:"noop" env:"NOOP" description:"If true fake board will be used"`
	Live   live.Options   `group:"live" namespace:"live" env-namespace:"LIVE"`
	Replay replay.Options `group:"replay" namespace:"replay" env-namespace:"REPLAY"`
	Snap   file.Options   `group:"snap" namespace:"snap" env-namespace:"SNAP"`
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
		prov lights.ControllerI
		err  error
	)

	prov = lights.NewTestController()

	if !opts.NoOp {
		prov, err = lights.NewController("/dev/i2c-0", opts.Boards)
		if err != nil {
			return fmt.Errorf("couldn't initiate controller for the boards: %w", err)
		}
	}

	g, ctx := errgroup.WithContext(appctx)

	snap := file.New(opts.Snap, afero.NewOsFs(), prov, internal.BuildingMap.Levels)

	lf := live.New(
		prov,
		internal.BuildingMap.Levels,
		opts.Live,
	)
	mf := manual.New(prov, internal.BuildingMap.Levels)
	rep := replay.New(afero.NewOsFs(), prov, opts.Replay)

	flowCtrl := flow.NewController(lf, mf, rep)

	srv, err := web.NewServer(prov, flowCtrl, snap, internal.BuildingMap.Levels)
	if err != nil {
		return fmt.Errorf("couln't initiate web server: %w", err)
	}

	g.Go(func() error { return srv.Listen(ctx, opts.Listen) })
	g.Go(func() error { return srv.NotifyViaSSE(ctx) })
	g.Go(func() error {
		errCh := flowCtrl.SubscribeToErrors()
		for {
			select {
			case err := <-errCh:
				slog.Error("Flow error: %v", err)
			case <-ctx.Done():
				return nil
			}
		}
	})

	g.Go(func() error { return shutdown.Receive(ctx) })

	slog.Info("Service started and listen", "addr", opts.Listen)
	err = g.Wait()

	if err != nil && errors.Is(err, shutdown.ErrGracefulShutdown) {
		slog.Info("Service shutted down successfuly")
		return nil
	}

	slog.Error("Service shut down unexpectedly: %v", err)

	return err
}
