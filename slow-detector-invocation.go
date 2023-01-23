/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// SlowDetectorInvocation is a container used by SlowDetectorCore
type SlowDetectorInvocation struct {
	sID       slowID
	threadID  ThreadID
	invoLabel string
	t0        time.Time
	stop      func(sdi *SlowDetectorInvocation, value ...time.Time)
	sd        *SlowDetectorCore
	tx        AtomicReference[time.Time]
}

// Stop ends an invocation part of SlowDetectorCore
func (sdi *SlowDetectorInvocation) Stop(value ...time.Time) {
	sdi.stop(sdi, value...)
}

// ThreadID returns the thread ID dor the thread invoking Start
func (sdi *SlowDetectorInvocation) ThreadID() (threadID ThreadID) {
	return sdi.threadID
}

// T0 returns the effective time of the invocation of Start
func (sdi *SlowDetectorInvocation) T0() (t0 time.Time) {
	return sdi.t0
}

// Label returns the label for this invocation
func (sdi *SlowDetectorInvocation) Label() (label string) {
	return sdi.invoLabel
}

// T0 returns the effective time of the invocation of Start
func (sdi *SlowDetectorInvocation) Time(t time.Time) (previousT time.Time) {
	addr := (*unsafe.Pointer)(unsafe.Pointer(&sdi.tx))
	var tp *time.Time
	if t.IsZero() {
		tp = (*time.Time)(atomic.LoadPointer(addr))
	} else {
		tp = (*time.Time)(atomic.SwapPointer(addr, unsafe.Pointer(&t)))
	}
	if tp != nil {
		previousT = *tp
	}
	return
}
