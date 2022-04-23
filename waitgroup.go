/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
)

/*
parl.WaitGroup is like a sync.Waitgroup that can be inspected.
The Waiting method returns the number of threads waited for.
parl.WaitGroup requires no initialization.
 var wg parl.WaitGroup
 wg.Add(1)
 …
 wg.Waiting()
*/
type WaitGroup struct {
	sync.WaitGroup
	lock  sync.Mutex
	adds  int
	dones int
}

func (wg *WaitGroup) Add(delta int) {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	wg.adds += delta
	wg.WaitGroup.Add(delta)
}

func (wg *WaitGroup) Done() {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	wg.dones++
	wg.WaitGroup.Done()
}

func (wg *WaitGroup) Counters() (adds int, dones int) {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	adds = wg.adds
	dones = wg.dones
	return
}

func (wg *WaitGroup) IsZero() (isZero bool) {
	adds, dones := wg.Counters()
	return adds == dones
}
