/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

type invokeTimerInvo struct {
	invokeTimer *InvokeTimer
	emID        emid
	t0          time.Time
	threadID    ThreadID
}

func newInvokeTimerInvo(invokeTimer *InvokeTimer) (invocation *invokeTimerInvo) {
	return &invokeTimerInvo{
		invokeTimer: invokeTimer,
		emID:        emidGenerator.ID(),
		t0:          time.Now(),
		threadID:    goID(),
	}
}

func (i *invokeTimerInvo) init() (invocation *invokeTimerInvo) {
	invocation = i
	i.invokeTimer.invoTimes.Put(i.emID, i)
	i.invokeTimer.ensureTimer()
	invos := i.invokeTimer.invos.Value()
	if i.invokeTimer.parallelism.Value(invos) {
		max, _ := i.invokeTimer.latency.Max()
		// callback for high parallelism warning
		i.invokeTimer.callback(ITParallelism, invos, max, invocation.threadID)
	}
	return
}

func (i *invokeTimerInvo) deferFunc() {
	i.invokeTimer.invoTimes.Delete(i.emID)
	i.invokeTimer.maybeCancelTimer()
	if duration := time.Since(i.t0); i.invokeTimer.latency.Value(duration) {
		max, _ := i.invokeTimer.parallelism.Max()
		// callback for slowness of completed task
		i.invokeTimer.callback(ITLatency, max, duration, i.threadID)
	}
}

// oldestFirst is an order function sorting the oldest invocation first
func (i *invokeTimerInvo) oldestFirst(a, b *invokeTimerInvo) (result int) {
	return a.t0.Compare(b.t0)
}
