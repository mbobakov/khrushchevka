//go:build manual

package live

import (
	"context"
	"io"
	"log/slog"
	"math/rand"
	"testing"
	"time"

	"github.com/mbobakov/khrushchevka/internal"
	"github.com/stretchr/testify/require"
)

// It's use lights matrix 3x3 with the following lights:
//
//	s - service light NoManLand
//	w - window light
//	n - no light
//	e - entrance light
//	1,2,3 - flat number
//
// then we will compare the calls to the provider with the expected calls
func TestLive_mainCycle(t *testing.T) {
	tests := []struct {
		name      string
		onGoing   map[int]bool
		mapping   [][]internal.Light
		wantCalls []map[string]bool // pin -> isON
	}{
		{
			// Map:
			//
			// s w(3) n(3) w(3)
			// s w(2) n(2) w(2)
			// e w(1) n(1) w(1)
			name:    "simple",
			onGoing: map[int]bool{1: true, 2: true}, // 3 flat will be selected
			mapping: [][]internal.Light{
				// ground floor
				{
					{Number: 0, Kind: internal.LightTypeServiceEntrance, Addr: internal.LightAddress{Board: 0, Pin: "0"}},
					{Number: 1, Kind: internal.LightTypeShortWindow, Addr: internal.LightAddress{Board: 0, Pin: "1"}},
					{Number: 1, Kind: internal.LightTypeWallStub},
					{Number: 1, Kind: internal.LightTypeShortWindow, Addr: internal.LightAddress{Board: 0, Pin: "2"}},
				},
				// 1st floor
				{
					{Number: 0, Kind: internal.LightTypeServiceNoManLand, Addr: internal.LightAddress{Board: 0, Pin: "3"}},
					{Number: 2, Kind: internal.LightTypeShortWindow, Addr: internal.LightAddress{Board: 0, Pin: "4"}},
					{Number: 2, Kind: internal.LightTypeWallStub},
					{Number: 2, Kind: internal.LightTypeShortWindow, Addr: internal.LightAddress{Board: 0, Pin: "5"}},
				},
				// 2nd floor
				{
					{Number: 0, Kind: internal.LightTypeServiceNoManLand, Addr: internal.LightAddress{Board: 0, Pin: "6"}},
					{Number: 3, Kind: internal.LightTypeShortWindow, Addr: internal.LightAddress{Board: 0, Pin: "7"}},
					{Number: 3, Kind: internal.LightTypeWallStub},
					{Number: 3, Kind: internal.LightTypeShortWindow, Addr: internal.LightAddress{Board: 0, Pin: "8"}},
				},
			},
			wantCalls: []map[string]bool{
				{"3": true},
				{"3": false},
				{"6": true},
				{"6": false},
				{"8": false},
				{"7": true},
				{"7": false},
				{"8": true},
				{"8": false},
				{"7": true},
				{"7": false},
				{"8": false},
				{"6": true},
				{"6": false},
				{"3": true},
				{"3": false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := []map[string]bool{}
			prov := &LightsControllerMock{
				SetFunc: func(addr internal.LightAddress, isON bool) error {
					calls = append(calls, map[string]bool{addr.Pin: isON})
					return nil
				},
			}

			l := New(prov, tt.mapping, Options{
				MaxDelay:   time.Minute, // select only once
				FlatTTL:    time.Second,
				ServiceTTL: 10 * time.Millisecond,
				MaxChanges: 3,
			}) // for determinism

			l.log = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			l.onGoing = tt.onGoing
			l.rand = rand.New(rand.NewSource(0))

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := l.mainCycle(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.wantCalls, calls)
		})
	}
}
