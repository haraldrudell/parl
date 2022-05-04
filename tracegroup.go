/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strings"
	"sync"
)

const (
	maxList    = 100
	skipFrames = 2
)

// parl.TraceGroup is an observable sync.Waitgroup.
//
// TraceGroup cannot be in parl because WaitAction imports goid
type TraceGroup struct {
	WaitGroup
	lock sync.Mutex
	list []WaitAction // behind lock
}

func (wg *TraceGroup) Add(delta int) {
	wg.WaitGroup.Add(delta)
	wg.action(delta, false)
}

func (wg *TraceGroup) Done() {
	wg.WaitGroup.Done()
	wg.action(0, true)
}

func (wg *TraceGroup) DoneBool() (isZero bool) {
	wg.WaitGroup.Done()
	wg.action(0, true)
	return wg.IsZero()
}

func (wg *TraceGroup) action(delta int, isDone bool) {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	if len(wg.list) == maxList {
		copy(wg.list, wg.list[1:])
	}
	wg.list = append(wg.list, *NewWaitAction(skipFrames, delta, isDone))
}

func (wg *TraceGroup) String() (s string) {
	adds, dones := wg.WaitGroup.Counters()
	s = Sprintf("%d(%d)", dones, adds)

	wg.lock.Lock()
	defer wg.lock.Unlock()

	if len(wg.list) == 0 {
		return
	}
	if len(wg.list) == 1 {
		return s + "\x20" + wg.list[0].String()
	}
	sL := make([]string, len(wg.list))
	for i, action := range wg.list {
		sL[i] = action.String()
	}
	return s + ":\n" + strings.Join(sL, "\n")
}
