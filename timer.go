/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"time"
)

const (
	tms = time.Millisecond
)

// Timer is a simple request timer
type Timer struct {
	Label string
	t0    time.Time
	d     time.Duration
}

// NewTimer gets a simple timer with duration or string output
func NewTimer(label string) (t *Timer) {
	t = &Timer{Label: label, t0: time.Now()}
	return
}

// End gets duration
func (t *Timer) End() (d time.Duration) {
	d = time.Since(t.t0)
	t.d = d
	return
}

// Endms gets tring with duration in ms
func (t *Timer) Endms() (ms string) {
	d := t.End()
	ms = fmt.Sprintf("%s: %s", t.Label, d.Round(tms).String())
	return
}
