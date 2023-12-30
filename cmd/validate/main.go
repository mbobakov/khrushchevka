package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mbobakov/khrushchevka/internal"
	"github.com/mbobakov/khrushchevka/internal/lights"
)

var (
	pins = []string{"A0", "A1", "A2", "A3", "A4", "A5", "A6", "A7", "B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7"}
)

type options struct {
	Delay    time.Duration `long:"delay" env:"DELAY" default:"3s" description:"Delay between go the then next light"`
	Boards   []uint8       `long:"boards" env:"BOARDS" default:"0x20,0x21,0x22,0x23,0x24,0x25" env-delim:"," description:"Boards to validate"`
	StartPin string        `long:"start-pin" env:"START_PIN" default:"A0" description:"Start Pin"`
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
	cntrl, err := lights.NewController("/dev/i2c-0", opts.Boards)
	if err != nil {
		return fmt.Errorf("couldn't initiate controller for the boards: %w", err)
	}

	for _, b := range opts.Boards {
		count := 0
		for {
			idx := count
			if idx >= len(pins) {
				idx = len(pins) - idx
			}
			if opts.StartPin == pins[idx] && idx > 0 {
				break
			}
			log.Printf("---\n Board: 0x%x Pin: %s is ON\n---\n", b, pins[idx])

			err := cntrl.Set(internal.LightAddress{
				Pin:   pins[idx],
				Board: b,
			}, true)
			if errors.Is(err, internal.ErrNoBoardConnected) {
				log.Printf("Board isn't connected!")
				continue
			}
			if err != nil {
				return fmt.Errorf("couldn't set pin to High: %w", err)
			}
			time.Sleep(opts.Delay) // sleep
			err = cntrl.Set(internal.LightAddress{
				Pin:   pins[idx],
				Board: b,
			}, false)
			if err != nil {
				return fmt.Errorf("couldn't set pin to Low: %w", err)
			}

			count++
		}

	}

	return nil
}
