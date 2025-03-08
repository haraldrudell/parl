/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

// SlowDetectorInvocationEnder is object able to end an invocation
type SlowDetectorInvocationEnder interface {
	// Stop ends an invocation created by SlowDetectorCore
	//	- invocation: the invocation object
	//	- timestamp: optional ending timestamp, default now
	Stop(invocation *SlowDetectorInvocation, timestamp ...time.Time)
}
