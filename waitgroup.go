/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
)

// parl.WaitGroup is like sync.Waitgroup with a Waiting method added.
// The waiting method returns the number of threads waiting
type WaitGroup struct {
	sync.WaitGroup
	lock    sync.Mutex
	waiting int
}

func (wg *WaitGroup) Add(delta int) {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	wg.waiting += delta
	wg.WaitGroup.Add(delta)
}

func (wg *WaitGroup) Done() {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	wg.waiting--
	wg.WaitGroup.Done()
}

func (wg *WaitGroup) Waiting() (waiting int) {
	wg.lock.Lock()
	defer wg.lock.Unlock()
	return wg.waiting
}

func (wg *WaitGroup) NoneWaiting() (noneWaiting bool) {
	return wg.Waiting() == 0
}
