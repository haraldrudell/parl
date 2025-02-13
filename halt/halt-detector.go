/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package halt

import (
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
)

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
	//	- Go garbage collector stop-the-world max pause is 1 ms.
	//		Typical pause is 100 μs.
	//		For monitoring the garbage collector, a tentative start is 100 μs.
	//	- [time.Sleep] is unreliable and cannot be used.
	//		Time.Sleep delays at least 1 ms on Linux
	//	- [runtime.Gosched] may delay over 1 s for 10k threads
	//	- [time.NewTimer] delays at least 374 ns
	//	- [time.NewTicker] is least troublesome sleep method.
	//		Fastest interval when 1 ns requested is 142.6 ns for single-thread
	//	- [time.Now] is 1 μs precision on macOS
	//	- 1 KiB allocation make([]byte, 1024) is 151.4 ns
	//	- context switch from one thread to another via lock is 224.7 ns
	//	- much be large enough so that slice allocation or thread-switch
	//		does not trigger a report
	//	- ticker, switch and allocation is 518.7 ns (142.6 + 151.4 + 224.7)
	//	- on macOS:
	//	- — smallest detectable halt is 11–29 µs
	//	- — 10 ms halt is exhibited within 1 s
	//	- — 30 ms halt is exhibited within 1 minute
	//	- on Linux:
	//	- — smallest detectable halt is 4.495 µs
	//	- — 2 ms halt is exhibited within 5 s
	//	- — 4 ms halt is exhibited within 2 minutes
	dynamicSleepTarget = time.Millisecond
)

// HaltDetector sends detected Go runtime execution halts on channel ch.
type HaltDetector struct {
	// the configured sleep interval default 1 ms
	interval parl.Atomic64[time.Duration]
	// the minimum halt duration that is reported, default 30 ms
	reportingThreshold parl.Atomic64[time.Duration]
	// isMonotonic means only larger halts than previously encountered
	// are reported
	//	- interval and reportingThreshold are dynamically adjusted
	//	- interval goes towards 1 ms
	isMonotonic atomic.Bool
	// unbound thread-safe slice receiving reports
	//	- closes on context cancel
	//	- value-channel to minimize allocations as opposed to pointers
	ch parl.AwaitableSlice[HaltReport]
}

// NewHaltDetector returns an object that sends detected Go runtime
// execution halts to an unbound thread-safe slice
//   - reportingThreshold: optional halt duration that is reported, default 30 ms
//   - — cannot be shorter than 30 ms
//   - — for shorter values, use [HaltDetector.SetThreshold]
//     -
//   - [HaltDetector.Thread] starts monitoring
//   - [HaltDetector.SetMonotonic] reports only increasing halts
//   - [HaltDetector.SetInterval] sets halt detector sleep period
//   - [HaltDetector.SetThreshold] sets inimum hgalt duration that is reported
//   - the unbound channel prevents consumer from holding up detection.
//     Eg. printing to macOS terminal window could be 6 seconds.
func NewHaltDetector(reportingThreshold ...time.Duration) (haltDetector *HaltDetector) {
	haltDetector = &HaltDetector{}
	var t time.Duration
	if len(reportingThreshold) > 0 {
		t = reportingThreshold[0]
	}
	if t < defaultReportingThreshold {
		t = defaultReportingThreshold
	}
	haltDetector.reportingThreshold.Store(t)
	haltDetector.interval.Store(defaultSleepInterval)
	return
}

// SetMonotonic reports only monotonically increasing halts
//   - useful to detect prevailing halts for a give architecture
//   - thread-safe
func (h *HaltDetector) SetMonotonic() {
	if !h.isMonotonic.Load() {
		h.isMonotonic.Store(true)
	}
}

// SetInterval reports only monotonically increasing halts
//   - sleep: how long to sleep prior to check for halt.
//     zero or negative is 1 ns to avoid panic
//   - for monotonic, sleep is dynamically increased
//     -
//   - to detect a given halt duration, sleep should be less than 50%
//   - for non-monotonic, values below 1 ms is high cpu load
//   - macOS halts are commonly 30 ms and can maximum do 30 µs
//   - Linux halts are commonly 5 ms and can maximum do 5 µs
//   - thread-safe
func (h *HaltDetector) SetInterval(sleep time.Duration) { h.interval.Store(sleep) }

// SetThreshold sets the shortest halt duration that is reported
//   - threshold: duration
//   - if non-monotonic, shold be at least 2× interval
//   - for monotonic, threshold is dynamically increased
//     -
//   - for macOS, less than 30 ms is high reporting volume
//   - for Linux, less than 5 ms is high reporting volume
//   - thread-safe
func (h *HaltDetector) SetThreshold(threshold time.Duration) { h.reportingThreshold.Store(threshold) }

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
	var interval = h.interval.Load()
	// avoid panic
	if interval < time.Nanosecond {
		interval = time.Nanosecond
	}
	// timeTicker runs a sleep interval
	var timeTicker = time.NewTicker(interval)
	defer timeTicker.Stop()

	var (
		done         = g.Context().Done()
		C            = timeTicker.C
		endSleep     = time.Now()
		haltReport   HaltReport
		threshold    = h.reportingThreshold.Load()
		reportNumber int
	)
	for {

		// sleep one interval then check halt duration
		//	- remember start time
		haltReport.T = endSleep
		select {
		case <-done:
			return // g context cancel return
		case <-C:
			endSleep = time.Now()
			haltReport.D = endSleep.Sub(haltReport.T)
		}
		// timer was triggered

		// monotonic case
		if h.isMonotonic.Load() {

			// adjust interval if monotonic
			//	- while interval shorter than dynamicSleepTarget
			//	- and duration is greater than interval
			if interval < dynamicSleepTarget &&
				haltReport.D > interval {
				// determine new interval
				if haltReport.D < dynamicSleepTarget {
					interval = haltReport.D
				} else {
					interval = dynamicSleepTarget
				}
				// reconfigure ticker
				timeTicker.Reset(interval)
			}

			// lesser durations never report
			if haltReport.D <= threshold {
				continue
				// equal duration only reports first time
			} else if haltReport.D == threshold {
				if reportNumber > 0 {
					continue
				}
			} else {
				// strictly greater: updates threshold and reports
				threshold = haltReport.D
			}

			// non-monotonic: simple report filtering
		} else if haltReport.D < threshold {
			continue // too short to be reported
		}

		// report
		reportNumber++
		haltReport.N = reportNumber
		h.ch.Send(haltReport)
	}
}

// Ch returns an iterable source for reports of Go runtime execution halts
//   - Ch closes on completion, ie. context cancel
func (h *HaltDetector) Ch() (ch parl.IterableAllSource[HaltReport]) { return &h.ch }
