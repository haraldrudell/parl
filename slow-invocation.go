/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

// SlowInvocation is consumer methods for [SlowDetectorInvocation]
//   - returned by [SlowDetector.Start]
type SlowInvocation interface {
	// Stop ends an invocation created [SlowDetector.Start]
	//   - timestamp: optional timestamp, default now
	Stop(timestamp ...time.Time)
	// Interval adds a timestamped label to an ongoing invocation
	//   - label: printable timestamp identifier “lsofComplete”
	//   - timestamp: optional timestamp, default now
	Interval(label string, timestamp ...time.Time)
}
