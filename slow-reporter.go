/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

// SlowReporter receives reports of slow or non-returning invocations
type SlowReporter interface {
	// Report receives reports for the slowest-to-date invocation
	// and non-return reports every minute
	//	- invocation: the invocation created by [SlowDetectorCode.Start]
	//	- didReturn true: the invocation has ended
	//	- didReturn false: the invocation is still in progress, ie. a non-return report
	//	- duration: the latency causing report
	Report(invocation *SlowDetectorInvocation, didReturn bool, duration time.Duration)
}
