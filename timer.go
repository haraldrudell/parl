/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"time"
)

var (
	// [NewTimer2] no fieldp
	NoTimerField *Timer
)

// Timer is a simple request timer
type Timer struct {
	// timer name “timer1”
	Label string
	// start timestamp
	t0 time.Time
	// duration upon invocation of End
	d Atomic64[time.Duration]
}

// NewTimer gets a simple timer with duration or string output
//   - label: printable name, default “timer1”
//   - [Timer.End] returns duration
//   - [Timer.Endms] returns ms string “timer1: 23ms”
//   - macOS precision: 1 μs, otherwise ns
func NewTimer(label string) (t *Timer) { return NewTimer2(NoTimerField, label) }

// NewTimer gets a simple timer with duration or string output
//   - fieldp: optional field pointer
//   - label: printable name, default “timer1”
//   - t0: optional start time, default time.Now
//   - [Timer.End] returns duration
//   - [Timer.Endms] returns ms string “timer1: 23ms”
//   - macOS precision: 1 μs, otherwise ns
func NewTimer2(fieldp *Timer, label string, t0 ...time.Time) (t *Timer) {
	if fieldp != nil {
		t = fieldp
		t.d.Store(0)
	} else {
		t = &Timer{}
	}
	if label != "" {
		t.Label = label
	} else {
		t.Label = fmt.Sprintf("timer%d", lastTimerID.Add(1))
	}
	var t0ToUse time.Time
	if len(t0) > 0 {
		t0ToUse = t0[0]
	}
	if t0ToUse.IsZero() {
		t0ToUse = time.Now()
	}
	t.t0 = t0ToUse
	return
}

// End gets duration
//   - t1: optional end time
//   - timer duration
//   - only first invocation is effective
//   - Thread-safe
func (t *Timer) End(t1 ...time.Time) (d time.Duration) {
	if d = t.d.Load(); d != 0 {
		return
	}
	var t1ToUse time.Time
	if len(t1) > 0 {
		t1ToUse = t1[0]
	}
	if t1ToUse.IsZero() {
		t1ToUse = time.Now()
	}
	d = t1ToUse.Sub(t.t0)
	if t.d.CompareAndSwap(0, d) {
		return // this thread won
	}
	// other thread won
	d = t.d.Load()
	return
}

// Endms gets string with duration in ms
//   - “timer1: 23ms”
//   - rounded to nearest ms
//   - Thread-safe
func (t *Timer) Endms() (ms string) {
	return fmt.Sprintf("%s: %s", t.Label, t.End().Round(roundMultiple))
}

const (
	// the duration rounding multiple [Timer.Endms] uses
	roundMultiple = time.Millisecond
)

// timer numbering 1…
var lastTimerID Atomic64[int]
