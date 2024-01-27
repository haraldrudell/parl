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
	// if new-function is invoked with zero default duration,
	// default default-duration is 1 second
	defaultDefaultDuration = time.Second
)

// ThreadSafeTimer is a timer that can be Reset by multiple threads. Improvements:
//   - Stop and Reset are thread-safe
//   - Reset includes Stop in a fail-safe sequence and can be invoked at any time by any thread
//   - NewThreadSafeTimer has a defaultDuration of 1 second if argument is missing, zero or negative
//   - Reset also uses defaultDuration if argument is zero or negative
//   - fields and method signatures are identical to [time.Timer]
//   - ThreadSafeTimer may be copied or stored in slices or maps
//   - —
//   - cost: 1 sync.Mutex, 1 int64, 2 pointers, 2 allocations
//   - [time.Timer] public methods and fields: C Reset Stop
//   - time.Timer Reset must be in an uninterrupted Stop-drain-Reset sequence or memory leaks result
type ThreadSafeTimer struct {
	// the default duration for Reset method
	defaultDuration time.Duration
	// the timer initially stopped
	*time.Timer
	// resetLock makes Stop, drain and Reset sequence atomic
	resetLock *sync.Mutex
}

// NewThreadSafeTimer returns a running timer with thread-safe Reset
//   - default defaultDuration is 1 second for defaultDuration missing, zero or negative
//   - defaultDuration is the default for Reset when Reset argument is zero or negative
//   - Reset can be invoked at any time without any precautions.
//   - — [time.Timer.Reset] has many conditions to avoid memory leaks
//   - Stop and Reset methods are thread-safe
//   - a timer must either expire or have Stop invoked to release resources
//   - if timer was created in same thread or obtained via synchronize before,
//     read of field C is thread-safe
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
//   - —
//   - thread-safety is obtained by making the Stop-drain-Reset sequence atomic
//   - unsynchronized Reset will cause memory leaks
func (t *ThreadSafeTimer) Reset(duration time.Duration) {
	if duration <= 0 {
		duration = t.defaultDuration
	}
	t.resetLock.Lock()
	defer t.resetLock.Unlock()

	// Reset should be invoked only on:
	//	- stopped or expired timers
	//	- with drained channels
	t.Timer.Stop()
	select {
	case <-t.Timer.C:
	default:
	}
	t.Timer.Reset(duration)
}

// Stop prevents the Timer from firing
func (t *ThreadSafeTimer) Stop() (wasRunning bool) {
	t.resetLock.Lock()
	defer t.resetLock.Unlock()

	return t.Timer.Stop()
}
