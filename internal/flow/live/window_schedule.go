package live

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type windowSchedule []struct {
	isOn     bool
	duration time.Duration
}

func (w windowSchedule) String() string {
	programToLog := strings.Builder{}
	for _, p := range w {
		programToLog.WriteString(fmt.Sprintf(":for %s is %t:", p.duration, p.isOn))
	}
	return programToLog.String()
}

func getWindowSchedule(rand *rand.Rand, ttl time.Duration, changes uint) windowSchedule {
	timeLeft := ttl
	changesLeft := changes
	isOn := []bool{true, false}[rand.Intn(2)]

	program := windowSchedule{}

	for changesLeft > 0 {
		durationRaw := timeLeft / time.Duration(changesLeft)

		duration := randomizeDuration(rand, durationRaw, 0.3)

		program = append(program, struct {
			isOn     bool
			duration time.Duration
		}{
			isOn:     isOn,
			duration: duration,
		})
		timeLeft -= duration
		changesLeft--
		isOn = !isOn
	}

	return program
}
