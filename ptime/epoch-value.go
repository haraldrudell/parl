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

func (ev *EpochValue) Get() (epoch Epoch) {
	return Epoch(atomic.LoadInt64((*int64)(ev)))
}

func (ev *EpochValue) Set(epoch Epoch) (oldEpoch Epoch) {
	return Epoch(atomic.SwapInt64((*int64)(ev), int64(epoch)))
}

func (ev *EpochValue) SetTime(t ...time.Time) (oldEpoch Epoch) {
	return Epoch(atomic.SwapInt64((*int64)(ev), int64(EpochNow(t...))))
}
