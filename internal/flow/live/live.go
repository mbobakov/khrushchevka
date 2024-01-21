package live

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/mbobakov/khrushchevka/internal"
	"golang.org/x/sync/errgroup"
)

type LightsController interface {
	Set(addr internal.LightAddress, isON bool) error
	Reset() error
}

const (
	name = "live"
)

// 1. once in maxDelay we are selecting randow flat
// 2. then send signal to enable service lights in stack for Service TTL
// 3. Livetime of random flat is ttl +/- 40% randomly
// 4. if ttl is reached, then switch off all lights in flat
// 5. During live time each window of the flat performing a program generated for the live time
// 6. Program contains <maxchanges> changes of light state with random delay between 0 and <maxInternalDelay>
//     max live time for the interval in program : <livetime>/<maxchanges> +/- 30%

type Options struct {
	MaxDelay   time.Duration `long:"max-delay" env:"MAX_DELAY" default:"30s" description:"max delay between flat selection"`
	FlatTTL    time.Duration `long:"flat-ttl" env:"FLAT_TTL" default:"5h" description:"flat live time"`
	ServiceTTL time.Duration `long:"service-ttl" env:"SERVICE_TTL" default:"20s" description:"service live time"`
	MaxChanges uint          `long:"max-changes" env:"MAX_CHANGES" default:"25" description:"max changes in flat per window"`
}

type Live struct {
	lights  LightsController
	mapping [][]internal.Light
	opts    Options
	done    chan struct{}
	log     *slog.Logger
	rand    *rand.Rand

	mu       sync.RWMutex
	isActive bool
}

//go:generate ../../../bin/moq -out mocks_test.go . LightsController
func New(l LightsController, mapping [][]internal.Light, opts Options) *Live {
	return &Live{
		lights:  l,
		opts:    opts,
		mapping: mapping,
		log:     slog.With("flow", name),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())), //no-lint: gosec
		done:    make(chan struct{}),
	}
}

func (l *Live) Name() string {
	return name
}

func (l *Live) serviceOn(ctx context.Context, sig <-chan struct{}, ttl time.Duration) error {
	allServiceLights := []internal.LightAddress{}
	for _, l := range l.mapping {
		for _, f := range l {
			if f.Kind == internal.LightTypeServiceNoManLand {
				allServiceLights = append(allServiceLights, f.Addr)
			}
		}
	}

	t := time.NewTimer(ttl)
	for {
		select {
		case <-sig:
			// switch on all service lights
			for _, v := range allServiceLights {
				err := l.lights.Set(v, true)
				if err != nil {
					return fmt.Errorf("couldn't switch on light '%v': %w", v, err)
				}
			}
			t.Reset(ttl)
		case <-ctx.Done():
			return nil
		case <-t.C:
			// switch off all service lights
			for _, v := range allServiceLights {
				err := l.lights.Set(v, false)
				if err != nil {
					return fmt.Errorf("couldn't switch off light '%v': %w", v, err)
				}
			}

		}
	}
}

func (l *Live) Start(ctx context.Context) error {
	l.log.Info("starting flow", slog.Any("opts", l.opts))
	l.done = make(chan struct{})
	l.mu.Lock()
	l.isActive = true
	l.mu.Unlock()

	// switch on entrance light
	l.log.Info("swithching off all lights but entrance")
	entrance := internal.Light{}
LOOP:
	for _, row := range l.mapping {
		for _, light := range row {
			if light.Kind == internal.LightTypeServiceEntrance {
				entrance = light
				break LOOP
			}
		}
	}

	err := l.lights.Reset()
	if err != nil {
		return fmt.Errorf("couldn't switch off lights: %w", err)
	}

	err = l.lights.Set(entrance.Addr, true)
	if err != nil {
		return fmt.Errorf("couldn't switch on entrance light '%v': %w", entrance.Addr, err)
	}

	return l.mainCycle(ctx)
}

func (l *Live) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isActive {
		l.log.Info("stopping flow")
		l.isActive = false
		close(l.done)
	}
}

func (l *Live) mainCycle(ctx context.Context) error {
	slog.Info("starting real life flow", slog.Any("opts", l.opts))
	var (
		serviceOnChan       = make(chan struct{})
		nextFlatSelectionIn = time.Duration(0)
		timer               = time.NewTimer(l.opts.MaxDelay)
		onGoing             = map[int]bool{}
		flats               = []int{}
	)
	// collect flats
	for _, l := range l.mapping {
		for _, w := range l {
			if w.Number != 0 {
				flats = append(flats, w.Number)
			}
		}
	}

	// make error group to wait for all goroutines
	group, ctx := errgroup.WithContext(ctx)

	// Start service lights
	group.Go(func() error { return l.serviceOn(ctx, serviceOnChan, l.opts.ServiceTTL) })

	for {
		l.log.Info("next flat selection in", slog.Duration("duration", nextFlatSelectionIn))

		timer.Reset(nextFlatSelectionIn)

		select {
		case <-ctx.Done():
			return nil
		case <-l.done:
			return nil
		case <-timer.C:
			if len(onGoing) == len(flats) {
				l.log.Info("all flats are busy, waiting for next flat selection")
				continue
			}

			selectedFlat := 0
			for {
				nominant := flats[l.rand.Intn(len(flats))]
				if onGoing[nominant] {
					continue
				}
				selectedFlat = nominant
				onGoing[selectedFlat] = true
				break
			}

			l.log.Info("selected flat", slog.Int("flat", selectedFlat))

			serviceOnChan <- struct{}{}

			// start the flat routine
			group.Go(func() error {
				err := l.flatCycle(ctx, group, selectedFlat)
				if err != nil {
					l.log.Error("couldn't process flat cycle", slog.Int("flat", selectedFlat), slog.Any("err", err))
					return fmt.Errorf("couldn't process flat cycle: %w", err)
				}

				delete(onGoing, selectedFlat)
				return nil
			})

			nextFlatSelectionIn = randomizeDuration(l.rand, l.opts.MaxDelay, 0.4)
		}
	}
}

func (l *Live) flatCycle(ctx context.Context, group *errgroup.Group, flat int) error {
	flatWindows := []internal.LightAddress{}

	for _, l := range l.mapping {
		for _, w := range l {
			if w.Number == flat && w.Kind != internal.LightTypeWallStub {
				flatWindows = append(flatWindows, w.Addr)
			}
		}

	}
	l.log.Info("flat cycle", slog.Int("flat", flat), slog.Duration("ttl", l.opts.FlatTTL))

	// Start flat live
	for _, fw := range flatWindows {
		schedule := getWindowSchedule(l.rand, l.opts.FlatTTL, l.opts.MaxChanges)
		l.log.Info("window schedule", slog.Int("flat", flat), slog.Any("addr", fw), slog.String("schedule", schedule.String()))

		fw := fw
		// Start window routine
		group.Go(func() error { return l.executeScheduleFor(ctx, schedule, fw) })
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-l.done:
			return nil
		}
	}

}

// executeScheduleFor executes the given schedule for a specific light address.
// It starts executing the program by setting the light on or off based on the schedule.
// It continues executing the program until the schedule is completed or the context is done.
// If the context is done or the execution is interrupted, it returns nil.
// If there is an error while switching the light, it returns an error with a formatted message.
func (l *Live) executeScheduleFor(ctx context.Context, schedule windowSchedule, addr internal.LightAddress) error {
	// Start executing program
	timer := time.NewTimer(l.opts.FlatTTL)
	for _, p := range schedule {
		timer.Reset(p.duration)
		err := l.lights.Set(addr, p.isOn)
		if err != nil {
			return fmt.Errorf("couldn't switch light '%v': %w", addr, err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-l.done:
			return nil
		case <-timer.C:
			continue
		}
	}

	return nil
}

// randomizeDuration randomize time <t> in the border of +/- <fluctuation * 100 > percent
func randomizeDuration(r *rand.Rand, t time.Duration, fluctuation float64) time.Duration {
	// Generate a random percentage within the fluctuation range
	randomPercentage := (r.Float64() * 2 * fluctuation) - fluctuation

	// Apply the fluctuation to the original duration
	return t + time.Duration(float64(t)*randomPercentage)
}
