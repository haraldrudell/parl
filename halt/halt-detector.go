/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package halt detects Go runtime execution halts.
package halt

import (
	"time"

	"github.com/haraldrudell/parl"
)

const (
	// timeSleepDuration is how long sleep is requested from the Go runtime time.Sleep function
	//	- needs to be about 10x shorter than the advertised minimum Go garbage-collector stop-the-world
	//		intervale, which is 1 ms
	timeSleepDuration         = 100 * time.Microsecond
	defaultReportingThreshold = 500 * time.Microsecond
)

// HaltDetector sends detected Go runtime execution halts on channel ch.
type HaltDetector struct {
	reportingThreshold time.Duration
	ch                 parl.NBChan[*HaltReport]
}

// HaltReport is a value object representing a detected Go runtime execution halt
type HaltReport struct {
	N int           // report number 1…
	T time.Time     // when halt started
	D time.Duration // halt duration
}

// NewHaltDetector returns an object that sends detected Go runtime execution halts on a channel
func NewHaltDetector(reportingThreshold ...time.Duration) (haltDetector *HaltDetector) {
	var t time.Duration
	if len(reportingThreshold) > 0 {
		t = reportingThreshold[0]
	}
	if t < defaultReportingThreshold {
		t = defaultReportingThreshold
	}
	return &HaltDetector{reportingThreshold: t}
}

// Thread detects execution halts and sends them on h.ch
func (h *HaltDetector) Thread(g0 parl.Go) {
	var err error
	defer g0.Register().Done(&err)
	defer parl.PanicToErr(&err)

	timeTicker := time.NewTicker(time.Millisecond)
	defer timeTicker.Stop()

	var done = g0.Context().Done()
	var elapsed time.Duration
	var t0 time.Time
	var t1 = time.Now()
	var reportingThreshold = h.reportingThreshold
	var n int
	var C = timeTicker.C
	for {

		// sleep
		t0 = t1
		select {
		case <-done:
			return // g0 context cancel return
		case <-C:
		}
		t1 = time.Now()
		elapsed = t1.Sub(t0)

		// report
		if elapsed >= reportingThreshold {
			n++
			h.ch.Send(&HaltReport{N: n, T: t0, D: elapsed})
		}
	}
}

// Ch returns a receive channel for reports of Go rutnime execution halts
//   - Ch never closes
func (h *HaltDetector) Ch() (ch <-chan *HaltReport) {
	return h.ch.Ch()
}
