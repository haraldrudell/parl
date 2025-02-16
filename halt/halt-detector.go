/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package halt

import (
	"time"

	"github.com/haraldrudell/parl"
)

const (
	// [NewHaltDetector2] halt detector reports only incrementally
	// longer halt durations
	MonotonicYes = true
)

// [NewHaltDetector2] control monotonicity [MonotonicYes]
type Monotonic bool

var (
	// [NewHaltDetector2] no pre-allocated field pointer is available
	NoHaltFieldp *HaltDetector
)

// HaltDetector sends detected Go runtime execution halts on channel ch.
type HaltDetector struct {
	// the configured sleep interval default 1 ms
	interval time.Duration
	// the minimum halt duration that is reported, default 30 ms
	reportingThreshold time.Duration
	// isMonotonic means only larger halts than previously encountered
	// are reported
	//	- interval and reportingThreshold are dynamically adjusted
	//	- interval goes towards 1 ms
	isMonotonic Monotonic
	// unbound thread-safe slice receiving reports
	//	- closes on context cancel
	//	- value-channel to minimize allocations as opposed to pointers
	ch parl.AwaitableSlice[HaltReport]
}

// NewHaltDetector returns an object that sends detected Go runtime
// execution halts to an unbound thread-safe slice
//   - reportingThreshold: optional halt duration that is reported, default 30 ms
//   - — cannot be shorter than 30 ms
//   - — for shorter values, use [NewHaltDetector2]
//     -
//   - [HaltDetector.Thread] starts monitoring
//   - the unbound channel prevents consumer from holding up detection.
//     Eg. printing to macOS terminal window could be 6 seconds.
func NewHaltDetector(reportingThreshold ...time.Duration) (haltDetector *HaltDetector) {
	var threshold time.Duration
	if len(reportingThreshold) > 0 {
		threshold = reportingThreshold[0]
	}
	if threshold < defaultReportingThreshold {
		threshold = defaultReportingThreshold
	}
	haltDetector = NewHaltDetector2(NoHaltFieldp, threshold, defaultSleepInterval)
	return
}

// NewHaltDetector2 returns an object that sends detected Go runtime
// execution halts to an unbound thread-safe slice
//   - fieldp: use as pre-allocated field pointer
//   - fieldp [NoHaltFieldp]: normal allocation
//   - reportingThreshold: minimum halt duration that is reported
//   - — values less than 30 ms may produce many reports
//   - — for monotonic, reportingThreshold is dynamically increased
//   - sleep: how long to sleep prior to check for halt
//   - — to detect a given halt duration, sleep should be 50% or less of reportingThreshold
//   - — for monotonic, sleep is dynamically increased
//   - isMonotonic [MonotonicYes]: reports only monotonically increasing halts
//   - — useful to detect prevailing halts for a give architecture
//   - — interval and reportingThreshold are dynamically adjusted
//   - — interval goes towards 1 ms
//     -
//   - [HaltDetector.Thread] starts monitoring
//   - for non-monotonic, threshold values below 1 ms is high cpu load
//   - sleep zero or negative is 1 ns to avoid panic
//   - macOS halts are commonly 30 ms and can shortest report 30 µs
//   - Linux halts are commonly 5 ms and can shortest report 5 µs
//   - the unbound channel prevents consumer from holding up detection.
//     Eg. printing to macOS terminal window could be 6 seconds.
func NewHaltDetector2(fieldp *HaltDetector, reportingThreshold, sleep time.Duration, isMonotonic ...Monotonic) (haltDetector *HaltDetector) {

	// set haltDetector
	if fieldp != nil {
		haltDetector = fieldp
		haltDetector.reportingThreshold = reportingThreshold
		haltDetector.interval = sleep
		haltDetector.ch = parl.AwaitableSlice[HaltReport]{}
	}
	if haltDetector == nil {
		haltDetector = &HaltDetector{
			reportingThreshold: reportingThreshold,
			interval:           sleep,
		}
	}

	// set monotonic
	if len(isMonotonic) > 0 {
		haltDetector.isMonotonic = isMonotonic[0]
	}

	return
}

// Ch returns an iterable source for reports of Go runtime execution halts
//   - Ch closes on completion, ie. context cancel
func (h *HaltDetector) Ch() (ch parl.IterableAllSource[HaltReport]) { return &h.ch }

// Thread detects execution halts and sends them on h.ch
//   - thread sleeps for 1 ms and every time that sleep
//     is longer than reportingThreshold, a report is issued
//   - thread runs until context cancel
//   - h.ch is unbound thread-safe slice
func (h *HaltDetector) Thread(g parl.Go) {
	var err error
	defer g.Register().Done(&err)
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)
	defer h.ch.EmptyCh()

	// the currently active interval
	// avoid panic
	if h.interval < time.Nanosecond {
		h.interval = time.Nanosecond
	}
	// timeTicker runs a sleep interval
	var timeTicker = time.NewTicker(h.interval)
	defer timeTicker.Stop()

	var (
		done         = g.Context().Done()
		C            = timeTicker.C
		endSleep     = time.Now()
		haltReport   HaltReport
		reportNumber int
	)
	for {

		// sleep one interval then check halt duration
		//	- remember start time
		haltReport.Timestamp = endSleep
		select {
		case <-done:
			return // g context cancel return
		case <-C:
			endSleep = time.Now()
			haltReport.Duration = endSleep.Sub(haltReport.Timestamp)
		}
		// timer was triggered

		// monotonic case
		if h.isMonotonic {

			// adjust interval if monotonic
			//	- while interval shorter than dynamicSleepTarget
			//	- and duration is greater than interval
			if h.interval < dynamicSleepTarget &&
				haltReport.Duration > h.interval {
				// determine new interval
				if haltReport.Duration < dynamicSleepTarget {
					h.interval = haltReport.Duration
				} else {
					h.interval = dynamicSleepTarget
				}
				// reconfigure ticker
				timeTicker.Reset(h.interval)
			}

			// lesser durations never report
			if haltReport.Duration <= h.reportingThreshold {
				continue
				// equal duration only reports first time
			} else if haltReport.Duration == h.reportingThreshold {
				if reportNumber > 0 {
					continue
				}
			} else {
				// strictly greater: updates threshold and reports
				h.reportingThreshold = haltReport.Duration
			}

			// non-monotonic: simple report filtering
		} else if haltReport.Duration < h.reportingThreshold {
			continue // too short to be reported
		}

		// report
		reportNumber++
		haltReport.Number = reportNumber
		h.ch.Send(haltReport)
	}
}

const (
	// defaultSleepInterval is a reasonable sleep time
	// conserving cpu capacity
	defaultSleepInterval = time.Millisecond
	// defaultReportingThreshold minimum
	//	- a new-function provided threshold duration cannot be less than this
	//	- should provide limited reporting on both Linux and macOS
	defaultReportingThreshold = 30 * time.Millisecond
	// dynamicSleepTarget is the interval used for halt detection
	//	- used for monotonic mode using dynamic interval configuration
	//	- the thread tries to get interval to at lest 1 ms to lower the cpu load
	dynamicSleepTarget = time.Millisecond
)
