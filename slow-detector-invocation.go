/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/haraldrudell/parl/ptime"
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
	lock      sync.Mutex
	intervals []Interval
}

type Interval struct {
	label string
	t     time.Time
}

// Stop ends an invocation part of SlowDetectorCore
func (sdi *SlowDetectorInvocation) Stop(value ...time.Time) {
	sdi.stop(sdi, value...)
}

// Stop ends an invocation part of SlowDetectorCore
func (sdi *SlowDetectorInvocation) Interval(label string, t ...time.Time) {
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	}
	if t0.IsZero() {
		t0 = time.Now()
	}

	sdi.lock.Lock()
	defer sdi.lock.Unlock()

	if label == "" {
		label = strconv.Itoa(len(sdi.intervals) + 1)
	}
	sdi.intervals = append(sdi.intervals, Interval{label: label, t: t0})
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

func (sdi *SlowDetectorInvocation) Intervals() (intervalStr string) {
	sdi.lock.Lock()
	defer sdi.lock.Unlock()

	if length := len(sdi.intervals); length > 0 {
		sList := make([]string, length)
		t0 := sdi.t0
		for i, ivl := range sdi.intervals {
			t := ivl.t
			sList[i] = ptime.Duration(t.Sub(t0)) + "\x20" + ivl.label
			t0 = t
		}
		intervalStr = "\x20" + strings.Join(sList, "\x20")
	}
	return
}
