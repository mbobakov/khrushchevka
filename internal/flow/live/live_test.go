package live

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/mbobakov/khrushchevka/internal"
	"github.com/stretchr/testify/require"
)

func TestLive_executeScheduleFor(t *testing.T) {
	tests := []struct {
		name      string
		schedule  windowSchedule
		addr      internal.LightAddress
		wantCalls []map[internal.LightAddress]bool // pin -> isON
		wantErr   bool
	}{
		{name: "simple",
			schedule: windowSchedule{
				{isOn: true, duration: 10 * time.Millisecond},
				{isOn: false, duration: 10 * time.Millisecond},
				{isOn: true, duration: 10 * time.Millisecond},
			},
			addr: internal.LightAddress{Board: 30, Pin: "A1"},
			wantCalls: []map[internal.LightAddress]bool{
				{internal.LightAddress{Board: 30, Pin: "A1"}: true},
				{internal.LightAddress{Board: 30, Pin: "A1"}: false},
				{internal.LightAddress{Board: 30, Pin: "A1"}: true},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := []map[internal.LightAddress]bool{}

			ctrlMoq := &LightsControllerMock{
				SetFunc: func(addr internal.LightAddress, isON bool) error {
					commands = append(commands, map[internal.LightAddress]bool{addr: isON})
					return nil
				},
			}

			l := &Live{
				lights: ctrlMoq,
				log:    slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
			}

			err := l.executeScheduleFor(context.Background(), tt.schedule, tt.addr)
			require.Equal(t, tt.wantErr, err != nil, "Live.executeScheduleFor() error = %v, wantErr %v", err, tt.wantErr)
			require.Equal(t, tt.wantCalls, commands)
		})
	}
}
