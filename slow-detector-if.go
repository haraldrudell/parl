/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

// SlowDetectorIf is interface provided by a slow detector
// to its slow-detector threads
type SlowDetectorIf interface {
	// Duration reports an invocation duration
	Duration(duration time.Duration) (isNewMax bool)
	// Report()
	SlowReporter
}

type SlowDetectorIf3 interface {
	SlowDetectorIf
	SlowDetectorInvocationEnder
}
