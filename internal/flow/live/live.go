package live

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/mbobakov/khrushchevka/internal"
	"golang.org/x/exp/maps"
)

type LightsController interface {
	Set(addr internal.LightAddress, isON bool) error
}

const (
	name = "live"
)

type Options struct {
	MaxDelay   time.Duration `long:"max-delay" env:"MAX_DELAY" default:"10s" description:"max delay between flat selection"`
	FlatTTL    time.Duration `long:"flat-ttl" env:"FLAT_TTL" default:"1m" description:"flat live time"`
	ServiceTTL time.Duration `long:"service-ttl" env:"SERVICE_TTL" default:"400ms" description:"service live time"`
	MaxChanges uint          `long:"max-changes" env:"MAX_CHANGES" default:"15" description:"max changes in flat per window"`
}

type Live struct {
	lights    LightsController
	mapping   [][]internal.Light
	flatsPool []int
	opts      Options
	done      chan struct{}
	log       *slog.Logger
	rand      *rand.Rand

	mu       sync.RWMutex
	onGoing  map[int]bool
	isActive bool

	servicePool []*internal.Light
	serviceLock sync.Mutex
}

//go:generate ../../../bin/moq -out mocks_test.go . LightsController
func New(l LightsController, mapping [][]internal.Light, opts Options) *Live {
	flats := map[int]struct{}{}
	serviceLevels := map[int]struct{}{}
	serviceLights := []*internal.Light{}
	for i, l := range mapping {
		for _, f := range l {
			f := f
			if f.Kind == internal.LightTypeServiceNoManLand {
				if _, ok := serviceLevels[i]; !ok {
					serviceLevels[i] = struct{}{}
					serviceLights = append(serviceLights, &f)
				}
			}
			if f.Number != 0 {
				flats[f.Number] = struct{}{}
			}
		}
		// add service empty light if floor has no service light
		if _, ok := serviceLevels[i]; !ok {

			serviceLevels[i] = struct{}{}
			serviceLights = append(serviceLights, &internal.Light{
				Kind: internal.LightTypeServiceNoManLand,
			})
		}
	}

	return &Live{
		lights:      l,
		opts:        opts,
		flatsPool:   maps.Keys(flats),
		mapping:     mapping,
		servicePool: serviceLights,
		log:         slog.With("flow", name),
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())), //no-lint: gosec
		onGoing:     make(map[int]bool),
		done:        make(chan struct{}),
	}
}

func (l *Live) Name() string {
	return name
}

func (l *Live) Start(ctx context.Context) error {
	l.log.Info("starting flow", slog.Any("opts", l.opts))
	l.done = make(chan struct{})
	l.mu.Lock()
	l.isActive = true
	l.mu.Unlock()

	// switch of all lights
	l.log.Info("swithching off all lights but entrance")
	for _, row := range l.mapping {
		for _, light := range row {
			if light.Kind == internal.LightTypeServiceEntrance {
				err := l.lights.Set(light.Addr, true)
				if err != nil {
					return fmt.Errorf("couldn't switch on entrance light '%v': %w", light.Addr, err)
				}
				continue
			}

			err := l.lights.Set(light.Addr, false)
			if err != nil {
				return fmt.Errorf("couldn't switch off light '%v': %w", light.Addr, err)
			}
		}
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
	t := time.NewTimer(l.opts.MaxDelay)
	first := true
	for {
		nextFlatSelectionIn := time.Duration(l.rand.Intn(int(l.opts.MaxDelay)))
		if nextFlatSelectionIn < time.Duration(float64(l.opts.MaxDelay)*0.6) {
			nextFlatSelectionIn = time.Duration(float64(l.opts.MaxDelay)*0.6 + float64(l.opts.MaxDelay)*0.4)
		}

		if first {
			nextFlatSelectionIn = 0
			first = false
		}

		l.log.Info("next flat selection in", slog.Duration("duration", nextFlatSelectionIn))
		t.Reset(nextFlatSelectionIn)

		select {
		case <-ctx.Done():
			return nil
		case <-l.done:
			return nil
		case <-t.C:
			tries := map[int]struct{}{}
			selectedFlat := 0
			l.mu.Lock()
			for {
				nominant := l.flatsPool[l.rand.Intn(len(l.flatsPool))]
				if l.onGoing[nominant] {
					tries[nominant] = struct{}{}
					if len(tries) == len(l.flatsPool) {
						l.log.Info("all flats are busy, waiting for next flat selection")
						break
					}
					continue
				}
				selectedFlat = nominant
				l.onGoing[selectedFlat] = true
				break
			}
			l.mu.Unlock()

			l.log.Info("selected flat", slog.Int("flat", selectedFlat))
			go func(flat int) {
				err := l.flatCycle(ctx, flat)
				if err != nil {
					l.log.Error("couldn't process flat cycle", slog.Int("flat", flat), slog.Any("err", err))
				}
			}(selectedFlat)
		}
	}
}

func (l *Live) flatCycle(ctx context.Context, flat int) error {
	wctx, cancel := context.WithTimeout(ctx, l.opts.FlatTTL)
	defer cancel()
	flatWindows := []internal.LightAddress{}
	floor := 0
	for i, l := range l.mapping {
		for _, w := range l {
			if w.Number == flat && w.Kind != internal.LightTypeWallStub {
				floor = i // no double deckers
				flatWindows = append(flatWindows, w.Addr)
			}
		}

	}
	l.log.Info("flat cycle", slog.Int("flat", flat), slog.Duration("ttl", l.opts.FlatTTL))

	// Start with elevation
	err := l.serviceProc(ctx, floor, true)
	if err != nil {
		return fmt.Errorf("couldn't process service proc for flat '%d': %w", flat, err)
	}

	// Start flat live
	for _, fw := range flatWindows {
		go func(addr internal.LightAddress) {
			err := l.windowCycle(wctx, flat, addr)
			if err != nil {
				l.log.Error("couldn't process window cycle", slog.Int("flat", flat), slog.Int("board", int(addr.Board)), slog.String("pin", addr.Pin), slog.Any("err", err))
			}
		}(fw)
	}

	defer func() {
		l.log.Info("flat cycle finished", slog.Int("flat", flat), slog.Duration("ttl", l.opts.FlatTTL))
		for _, w := range flatWindows {
			err := l.lights.Set(w, false)
			if err != nil {
				l.log.Error("couldn't switch off light", slog.Int("flat", flat), slog.Int("board", int(w.Board)), slog.String("pin", w.Pin), slog.Any("err", err))
			}
		}

		err = l.serviceProc(ctx, floor, false)
		if err != nil {
			l.log.Error("couldn't perform service proc", slog.Int("flat", flat))

		}
		l.mu.Lock()
		delete(l.onGoing, flat)
		l.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-l.done:
			return nil
		case <-wctx.Done():
			return nil
		}
	}

}

func (l *Live) windowCycle(ctx context.Context, flat int, addr internal.LightAddress) error {
	type period struct {
		isOn     bool
		duration time.Duration
	}

	program := []period{}

	timeLeft := l.opts.FlatTTL
	changesLeft := l.opts.MaxChanges
	isOn := []bool{true, false}[l.rand.Intn(2)]

	for changesLeft > 0 {
		durationRaw := timeLeft / time.Duration(changesLeft)

		duration := time.Duration(
			float64(durationRaw)*0.7 +
				float64(l.rand.Intn(1000))/1000.0*
					float64(durationRaw)*
					0.3*
					[]float64{-1, 1}[l.rand.Intn(2)], // +/-
		)

		program = append(program, period{
			isOn:     isOn,
			duration: duration,
		})
		timeLeft -= duration
		changesLeft--
		isOn = !isOn
	}

	programToLog := strings.Builder{}
	for _, p := range program {
		programToLog.WriteString(fmt.Sprintf(":for %s is %t:", p.duration, p.isOn))
	}
	l.log.Info("program for the flat window",
		slog.Int("flat", flat),
		slog.Int("board", int(addr.Board)),
		slog.String("pin", addr.Pin),
		slog.Any("program", programToLog.String()),
	)
	// Start executing program
	timer := time.NewTimer(l.opts.FlatTTL)
	for _, p := range program {
		timer.Reset(p.duration)

		select {
		case <-ctx.Done():
			return nil
		case <-l.done:
			return nil
		case <-timer.C:
			err := l.lights.Set(addr, !p.isOn)
			if err != nil {
				return fmt.Errorf("couldn't switch light '%v': %w", addr, err)
			}
		}
	}

	return nil
}

func (l *Live) serviceProc(ctx context.Context, level int, directionToUp bool) error {
	l.serviceLock.Lock()
	defer l.serviceLock.Unlock()
	l.log.Info("service proc", slog.Int("level", level), slog.Bool("directionToUp", directionToUp))
	currentPos := 0
	end := level
	n := 1
	if !directionToUp {
		currentPos = level
		end = 0
		n = -1
	}
	timer := time.NewTimer(l.opts.ServiceTTL)
	for {
		timer.Reset(l.opts.ServiceTTL)

		if currentPos < 0 || currentPos >= len(l.servicePool) {
			break
		}
		light := l.servicePool[currentPos]

		if len(light.Addr.Pin) == 0 {
			currentPos += n
			continue
		}
		err := l.lights.Set(light.Addr, true)
		if err != nil {
			return fmt.Errorf("couldn't switch on light '%v': %w", light.Addr, err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-l.done:
			return nil
		case <-timer.C:
		}

		err = l.lights.Set(light.Addr, false)
		if err != nil {
			return fmt.Errorf("couldn't switch off light '%v': %w", light.Addr, err)
		}

		if currentPos == end {
			break
		}
		currentPos += n
	}

	return nil
}

// 1. once in maxDelay we are selecting randow flat
// 2. then sequence of service lights in stack until floor is reached (block service lights for it)
// 3. Livetime of random flat is ttl +/- 20% randomly
// 4. if ttl is reached, then switch off all lights in flat
// 5. sequence of service lights in stack until ground floor is reached (block service lights for it)
// 6. During live time each window of the flat performing a program generated for the live time
// 7. Program contains <maxchanges> changes of light state with random delay between 0 and <maxInternalDelay>
//     max live time for the interval in program : <livetime>/<maxchanges> +/- 30%
