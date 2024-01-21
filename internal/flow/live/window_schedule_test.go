package live

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_getWindowSchedule(t *testing.T) {
	rand := rand.New(rand.NewSource(0))
	tests := []struct {
		name    string
		ttl     time.Duration
		changes uint
		want    windowSchedule
	}{
		{name: "simple",
			ttl: 5 * time.Second, changes: 5,
			want: windowSchedule{
				{isOn: true, duration: 846979052},
				{isOn: false, duration: 1135408682},
				{isOn: true, duration: 736907256},
				{isOn: false, duration: 1049754149},
				{isOn: true, duration: 1075467316},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := getWindowSchedule(rand, tt.ttl, tt.changes)

			require.Equal(t, tt.want, got)

		})
	}
}
