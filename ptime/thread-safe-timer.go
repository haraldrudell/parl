/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync"
	"time"
)

const (
	defaultDefaultDuration = time.Second
)

// ThreadSafeTimer is a timer that is used by multiple threads
//   - ThreadSafeTimer is a timer with Reset method thread-safe
//   - —
//   - [time.Timer] public methods and fields: C Reset Stop
//   - signatures identical to [time.Timer]
type ThreadSafeTimer struct {
	// the default duration for Reset method
	defaultDuration time.Duration
	// the timer initially stopped
	*time.Timer
	// resetLock makes Stop, drain and Reset sequence atomic
	resetLock *sync.Mutex
}

// NewThreadSafeTimer returns a timer with thread-safe Reset
//   - timer starts running
//   - default default duration is 1 second
//   - —
//   - promoted fields and methods: C Stop
func NewThreadSafeTimer(defaultDuration ...time.Duration) (timer *ThreadSafeTimer) {

	// determine default duration
	var d time.Duration
	if len(defaultDuration) > 0 {
		d = defaultDuration[0]
	}
	if d <= 0 {
		d = defaultDefaultDuration // 1 second
	}

	return &ThreadSafeTimer{
		defaultDuration: d,
		Timer:           time.NewTimer(d),
		resetLock:       &sync.Mutex{},
	}
}

// Reset is thread-safe timer reset
//   - has default duration for duration == 0
//   - works with concurrent channel read
//   - works with concurrent timer.Stop
//   - supports functional chaining
//   - —
//   - thread-safety is obtained by making the Stop-drain-Reset sequence atomic
func (t *ThreadSafeTimer) Reset(duration time.Duration) {
	if duration <= 0 {
		duration = t.defaultDuration
	}

	var timer0 = t.Timer
	var C = timer0.C
	t.resetLock.Lock()
	defer t.resetLock.Unlock()

	// Reset should be invoked only on:
	//	- stopped or expired timers
	//	- with drained channels
	timer0.Stop()
	select {
	case <-C:
	default:
	}
	timer0.Reset(duration)
}
