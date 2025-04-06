/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

// SlowDetectorInvoActions can report invocation durations and
// prompt reports: Duration() Report()
//   - interface provided by a slow detector
//     to its slow-detector threads
type SlowDetectorInvoActions interface {
	// Duration reports an invocation duration
	//	- duration: measured duration
	//	- isNewMax true: this is invocation is a progressive max
	//		for this slow detector
	//	- —
	//	- invoked periodically by the monioring thread
	//		during the invocation life
	//	- a progressive max detected by Duration prompts
	//		report output via Report
	//   - also records all-time max for the slow-detector
	Duration(duration time.Duration) (isNewMax bool)
	// [SlowReporter.Report]:
	// Report receives reports for the slowest-to-date invocation
	// and non-return reports every minute
	//	- invocation: the invocation created by [SlowDetectorCode.Start]
	//	- didReturn DidReturnYes: the invocation has ended
	//	- didReturn DidReturnNo: the invocation is still in progress, ie. a non-return report
	//	- duration: the latency causing report
	SlowReporter
}

// SlowDetectorIf can report invocation durations,
// prompt reports and
// end invocations:
// Duration() Report() Stop()
//   - interface provided by a slow detector
//     to its slow-detector threads
type SlowDetectorInvoActionsStop interface {
	// Duration() Report()
	SlowDetectorInvoActions
	// Stop ends an invocation created by [SlowDetectorCore.Start]
	//	- invocation: the invocation object returned by Start
	//	- timestamp: optional ending timestamp, default now
	Stop(invocation *SlowDetectorInvocation, timestamp ...time.Time)
}
