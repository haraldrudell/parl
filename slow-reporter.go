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
	//	- didReturn DidReturnYes: the invocation has ended
	//	- didReturn DidReturnNo: the invocation is still in progress, ie. a non-return report
	//	- duration: the latency causing report
	Report(invocation *SlowDetectorInvocation, didReturn DidReturn, duration time.Duration)
}

const (
	// Report: the invocation did return
	DidReturnYes DidReturn = iota + 1
	// Report: the invocation has yet to return
	DidReturnNo
)

// Report return flag: [DidReturnNo] []DidReturnYes
type DidReturn uint8
