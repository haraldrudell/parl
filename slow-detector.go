/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/ptime"
)

const (
	defaultThreshold = 100 * time.Millisecond
	newFrames        = 1
	ensureFrames     = 2
	slowScanPeriod   = time.Second
	slowScanInteval  = time.Minute
)

type slowID uint64

var slowIDGenerator UniqueIDTypedUint64[slowID]
var slowMap = pmaps.NewRWMap[slowID, *SlowDetector]()
var hasThread AtomicBool

// SlowDetector measures latency via Start-Stop invocations and prints
// max latency values above threshold to stderr. Thread-safe
type SlowDetector struct {
	max AtomicMax[time.Duration]
	t0  AtomicReference[time.Time]

	IDlock      sync.Mutex
	ID          slowID
	printf      PrintfFunc
	label       string
	threadID    ThreadID
	threadPrint time.Time
}

// NewSlowDetector returns a Start-Stop variable detecting slowness
//   - default label is code location of caller
//   - default threshold is 100 ms
//   - if threshold is 0, all max-slowness invocations are printed
//   - output is to stderr
func NewSlowDetector(label string, threshold ...time.Duration) (slowDetector *SlowDetector) {
	if label == "" {
		label = pruntime.NewCodeLocation(newFrames).Short()
	}
	var threshold0 time.Duration
	if len(threshold) > 0 {
		threshold0 = threshold[0]
	} else {
		threshold0 = defaultThreshold
	}
	return &SlowDetector{
		ID:     slowIDGenerator.ID(),
		label:  label,
		max:    NewAtomicMax(threshold0),
		printf: Log,
	}
}

// NewSlowDetector returns a Start-Stop variable detecting slowness
//   - default label is code location of caller
//   - default threshold is 100 ms
//   - if threshold is 0, all max-slowness invocations are printed
//   - output is the provided printf function
func NewSlowDetectorPrintf(label string, printf PrintfFunc, threshold ...time.Duration) (slowDetector *SlowDetector) {
	if printf == nil {
		panic(perrors.NewPF("printf cannot be nil"))
	}
	slowDetector = NewSlowDetector(label, threshold...)
	slowDetector.printf = printf
	return
}

func (sd *SlowDetector) Start() {
	sd.ensureID()
	t0 := time.Now()
	sd.t0.Put(&t0)
	sd.store()
}

func (sd *SlowDetector) Stop() {
	slowMap.Delete(sd.ID)
	var t0 time.Time
	if timep := sd.t0.Get(); timep == nil {
		panic(perrors.NewPF("no preceding Start invocation"))
	} else {
		t0 = *timep
	}
	if duration := time.Since(t0); sd.max.Value(duration) {
		sd.print(duration, false)
	}
}

func (sd *SlowDetector) store() {
	slowMap.Put(sd.ID, sd)
	if hasThread.Set() {
		go slowScanThread()
	}
}

func (sd *SlowDetector) ensureID() {
	sd.IDlock.Lock()
	defer sd.IDlock.Unlock()

	sd.threadID = goID()
	sd.threadPrint = time.Time{}
	if sd.ID != 0 {
		return
	}

	sd.ID = slowIDGenerator.ID()
	if sd.printf == nil {
		sd.printf = Log
	}
	if sd.label == "" {
		sd.label = pruntime.NewCodeLocation(ensureFrames).Short()
	}
}

func (sd *SlowDetector) print(duration time.Duration, isThread bool) {
	sd.IDlock.Lock()
	defer sd.IDlock.Unlock()

	var pStr string
	now := time.Now()
	if isThread {
		if now.Sub(sd.threadPrint) < slowScanInteval {
			return // only do thread-prints once a minute
		} else {
			sd.threadPrint = now
			pStr = " in progress…"
		}
	}

	var threadStr string
	if sd.threadID.IsValid() {
		threadStr = " threadID: " + sd.threadID.String()
	}

	sd.printf("%s max duration: %s%s%s", sd.label, ptime.Duration(duration), threadStr, pStr)
}

func slowScanThread() {
	Recover(Annotation(), nil, Infallible)

	ticker := time.NewTicker(slowScanPeriod)
	defer ticker.Stop()

	C := ticker.C
	for {
		<-C
		for _, sd := range slowMap.List() {
			var duration time.Duration
			if tp := sd.t0.Get(); tp == nil {
				continue
			} else {
				duration = time.Since(*tp)
			}
			if sd.max.Value(duration) {
				sd.print(duration, true)
			}
		}
	}
}
