/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// HaltDetector sends detected Go runtime execution halts on channel ch.
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
//   - HaltDetector takes average 50 μs max 100 μs to shut down
//     since time.Sleep cannot be aborted.
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
	defer parl.Recover(parl.Annotation(), &err, parl.NoOnError)

	var contextErr = g0.Context().Err
	var elapsed time.Duration
	var t0 time.Time
	var t1 = time.Now()
	var reportingThreshold = h.reportingThreshold
	var n int
	for {

		// sleep
		t0 = t1
		time.Sleep(timeSleepDuration)
		t1 = time.Now()
		elapsed = t1.Sub(t0)

		// report
		if elapsed >= reportingThreshold {
			n++
			h.ch.Send(&HaltReport{N: n, T: t0, D: elapsed})
		}

		// check for cancel
		if contextErr() != nil {
			return // g0 context cancel return
		}
	}
}

// Ch returns a receive channel for reports of Go rutnime execution halts
//   - Ch never closes
func (h *HaltDetector) Ch() (ch <-chan *HaltReport) {
	return h.ch.Ch()
}
