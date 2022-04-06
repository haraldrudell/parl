/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

type OnceChan struct {
	lock   sync.Mutex
	ch     chan struct{}
	isDone bool
}

func (oc *OnceChan) Done() (ch <-chan struct{}) {
	ch, _ = oc.get(false)
	return
}

func (oc *OnceChan) IsDone() (isDone bool) {
	_, isDone = oc.get(false)
	return
}

func (oc *OnceChan) Cancel() {
	oc.get(true)
}

func (oc *OnceChan) get(setCancel bool) (ch chan struct{}, isDone bool) {
	oc.lock.Lock()
	defer oc.lock.Unlock()

	ch = oc.ch
	isDone = oc.isDone
	if ch == nil {
		ch = make(chan struct{})
		oc.ch = ch
	}
	if setCancel && !isDone {
		isDone = true
		oc.isDone = true
		close(ch)
	}
	return
}
