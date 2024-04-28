/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import "time"

// Alert impelments a renewable alert timer. not thread-safe
type Alert struct {
	time.Duration
	// Timer.C can be awaited for any trigging of the timer
	//	- Timer.C never closes and only sends a single value on trig
	//	- Timer is reinitialized on each Start invocation
	*time.Timer
	// Fired may be set to true by consumer on Timer.C trig
	//	- doing so saves a [time.Timer.Stop] invocation on next [Alert.Stop]
	// true if the timer trigged after the last Start invocation prior to any Stop invocation
	Fired bool
}

// NewAlert allocates a renewable alert time of specific duration
//   - [Alert.Start] initializes a new period regardless of state
//   - [Alert.Stop] releases resources regardless of state. [Alert.Stop] is idempotent
//   - if [Alert.Start] was invoked, a subsequent [Alert.Stop] is required to release resources
//   - after Start, [Alert.Timer.C] can be awaited for any trigging of the timer
//   - [Alert.Timer.C] never closes and only sends a single value on trig
//   - [Alert.Fired] is true if Start was invoked and the timer trigged prior to Stop
func NewAlert(d time.Duration) (al *Alert) {
	al = &Alert{Duration: d}
	return
}

// Start initializes a new alert period
//   - can be invoked at any time. Not thread-safe
func (al *Alert) Start() {
	if al.Timer != nil && !al.Fired {
		al.Timer.Stop()
	}
	al.Timer = time.NewTimer(al.Duration)
	al.Fired = false
}

// Stop releases resources associated with this alert
//   - idempotent, can be invoked at any time. Not thread-safe
func (al *Alert) Stop() {
	if al.Timer != nil && !al.Fired {
		al.Timer.Stop()
	}
	al.Timer = nil
}
