/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/sets"
)

const (
	SlowDefault slowType = iota
	SlowOwnThread
	SlowShutdownThread

	slowScanPeriod = time.Second
)

// shared SlowDetectorThread for SlowDefault threads
var slowDetectorThread SlowDetectorThread

type SlowDetectorThread struct {
	slowTyp         slowType
	nonReturnPeriod time.Duration
	slowMap         pmaps.RWMap[slowID, *SlowDetectorInvocation]
	hasThread       AtomicBool

	slowLock sync.Mutex
	goGen    GoGen
	cancelGo func()
}

func NewSlowDetectorThread(slowTyp slowType, nonReturnPeriod time.Duration, goGen GoGen) (sdt *SlowDetectorThread) {
	if goGen == nil {
		panic(perrors.NewPF("goGen cannot be nil"))
	}

	// dedicated thread case
	if slowTyp != SlowDefault {
		return &SlowDetectorThread{
			slowTyp:         slowTyp,
			nonReturnPeriod: nonReturnPeriod,
			slowMap:         *pmaps.NewRWMap2[slowID, *SlowDetectorInvocation](),
			goGen:           goGen,
		}
	}

	sdt = &slowDetectorThread
	sdt.slowLock.Lock()
	defer sdt.slowLock.Unlock()

	if sdt.goGen != nil {
		return // slowDetectorThread already initialized return
	}

	// slowDetectorThread initialization
	sdt.slowTyp = slowTyp
	sdt.nonReturnPeriod = nonReturnPeriod
	sdt.slowMap = *pmaps.NewRWMap2[slowID, *SlowDetectorInvocation]()
	sdt.goGen = goGen

	return
}

func (sdt *SlowDetectorThread) Start(sdi *SlowDetectorInvocation) {

	// store in map
	sdt.slowMap.Put(sdi.sID, sdi)

	if !sdt.hasThread.Set() {
		return // thread already running return
	}

	// launch thread
	subGo := sdt.goGen.SubGo()
	g0 := subGo.Go()
	go sdt.thread(g0)
	if sdt.slowTyp != SlowShutdownThread {
		return // thread is not to be shutdown return
	}

	// save cancel method
	sdt.slowLock.Lock()
	defer sdt.slowLock.Unlock()

	sdt.cancelGo = subGo.Cancel
}

func (sdt *SlowDetectorThread) Stop(sdi *SlowDetectorInvocation) {

	// remove from map
	sdt.slowMap.Delete(sdi.sID, parli.MapDeleteWithZeroValue)

	if sdt.slowMap.Length() > 0 || sdt.slowTyp != SlowShutdownThread {
		return // not to be shutdown or not to be shutdown now return
	}

	sdt.cancelGo()
}

func (sdt *SlowDetectorThread) thread(g0 Go) {
	var err error
	defer g0.Register("SlowDetectorThread" + goID().String()).Done(&err)
	defer Recover(Annotation(), &err, NoOnError)

	ticker := time.NewTicker(slowScanPeriod)
	defer ticker.Stop()

	var C <-chan time.Time = ticker.C
	var done <-chan struct{} = g0.Context().Done()
	var t time.Time
	for {
		select {
		case <-done:
			return // context cancelled return
		case t = <-C:
		}

		// check all invocations for non-return
		for _, sdi := range sdt.slowMap.List() {
			// duration is how long the invocation has been in progress
			duration := t.Sub(sdi.t0)
			if duration < 0 {
				// if t coming from the ticker was delayed,
				// then t may be a time in the past,
				// so early that sdi.t0 is after t
				continue // ignore negative durations
			}
			sd := sdi.sd
			sd.alwaysMax.Value(duration)
			if sd.max.Value(duration) {
				// it is a new max, check whether nonReturnPeriod has elapsed
				if tLast := sdi.Time(time.Time{}); tLast.IsZero() || t.Sub(tLast) >= sdt.nonReturnPeriod {

					// store new nonReturnPeriod start
					sdi.Time(t)
					sd.callback(sdi, false, duration)
				}
			}
		}
	}
}

type slowType uint8

func (st slowType) String() (s string) {
	return slowTypeSet.StringT(st)
}

var slowTypeSet = sets.NewSet(sets.NewElements[slowType](
	[]sets.SetElement[slowType]{
		{ValueV: SlowDefault, Name: "sharedThread"},
		{ValueV: SlowOwnThread, Name: "ownThread"},
		{ValueV: SlowShutdownThread, Name: "shutdownThread"},
	}))
