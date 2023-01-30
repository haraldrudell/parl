/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// EpochValue is a timestamp with Thread-Safe atomic access.
package ptime

import (
	"sync/atomic"
	"time"
)

// EpochValue is a timestamp with Thread-Safe atomic access.
type EpochValue Epoch

// Get returns the current Epoch value
//   - zero epoch is returned as zero-time Time.IsZero, time.Time{}
//   - to get time.Time: epochValue.Get().Time()
//   - to check if non-zero: epochValue.Get().IsValid()
func (ev *EpochValue) Get() (epoch Epoch) {
	return Epoch(atomic.LoadInt64((*int64)(ev)))
}

// Set updates the Epoch value returning the old value
//   - 0 is the zero-value
func (ev *EpochValue) Set(epoch Epoch) (oldEpoch Epoch) {
	return Epoch(atomic.SwapInt64((*int64)(ev), int64(epoch)))
}

// SetTime updates the Epoch value to a time.Time value returning the old Epoch value
//   - default time value is time.Now()
//   - time.IsZero or time.Time{} is the zero-value
func (ev *EpochValue) SetTime(t ...time.Time) (oldEpoch Epoch) {
	return Epoch(atomic.SwapInt64((*int64)(ev), int64(EpochNow(t...))))
}
