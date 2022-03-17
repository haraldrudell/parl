/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parltime

import "time"

// Alert impelemnts a renewable alert timer
type Alert struct {
	time.Duration
	*time.Timer
	Fired bool
}

// NewAlert allocates a renewable alert time
func NewAlert(d time.Duration) (al *Alert) {
	al = &Alert{Duration: d}
	return
}

// Start initializes a new alert period
func (al *Alert) Start() {
	if al.Timer != nil && !al.Fired {
		al.Timer.Stop()
	}
	al.Timer = time.NewTimer(al.Duration)
	al.Fired = false
}

// Stop releases resources associated with this alert
func (al *Alert) Stop() {
	if al.Timer != nil && !al.Fired {
		al.Timer.Stop()
	}
	al.Timer = nil
}
