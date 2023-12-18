/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"context"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

const (
	wUnwatched watchState = iota
	wWatching
	wWatchFailed
)

type watchState uint8

type Iterator struct {
	path       string
	done       <-chan struct{}
	watcher    Watcher
	watchState parl.Atomic32[watchState]
	err        atomic.Pointer[error]
	isEvent    parl.CyclicAwaitable

	eventLock sync.Mutex
	events    pslices.Shifter[*WatchEvent]
}

// NewIterator returns a file-system watch-event iterator
//
// Usage:
//
// iterator = NewIterator(watchfs.WatchOpAll, watchfs.NoIgnores)
func NewIterator(path string, filter Op, ignores *regexp.Regexp, ctx context.Context) (iterator iters.Iterator[*WatchEvent]) {
	i := Iterator{
		path:    path,
		done:    ctx.Done(),
		isEvent: *parl.NewCyclicAwaitable(),
	}
	i.watcher = *NewWatcher(filter, ignores, i.receiveWatchEvent, i.errFn)
	return iters.NewFunctionIterator(i.iteratorFunction)
}

func (i *Iterator) iteratorFunction(isCancel bool) (fsEvent *WatchEvent, err error) {

	//handle cancel and error
	if isCancel {
		i.watcher.Shutdown()
	}
	err = i.error() // collect any Shutdown errors
	if isCancel || err != nil {
		return // cancel or error return
	}

	// ensure Watching
	if err = i.ensureWatch(); err != nil {
		return
	}

	// get or wait for file-system event, context cancel or error
	var isDone bool
	if fsEvent, isDone, err = i.event(); err != nil {
		return // recent error return
	}
	if !isDone {
		return // has event return
	}
	// it’s context cancel

	// shutdown watcher
	i.watcher.Shutdown()
	if err = i.error(); err != nil {
		return
	}
	err = parl.ErrEndCallbacks // signal cancel
	return
}

func (i *Iterator) ensureWatch() (err error) {

	// fast check outside lock
	var doWatch bool
	if doWatch, err = i.stateCheck(); !doWatch {
		return
	}
	i.eventLock.Lock()
	defer i.eventLock.Unlock()

	if doWatch, err = i.stateCheck(); !doWatch {
		return
	}
	if err = i.watcher.Watch(i.path); err == nil {
		i.watchState.Store(wWatching)
		return
	}
	i.watchState.Store(wWatchFailed)
	i.errFn(err)
	i.watcher.Shutdown()
	err = i.error()

	return
}

func (i *Iterator) stateCheck() (doWatch bool, err error) {
	switch i.watchState.Load() {
	case wUnwatched:
		doWatch = true
	case wWatchFailed:
		err = perrors.NewPF("Watch failed")
	}
	return
}

// receiveWatchEvent receives events from the watcher
func (i *Iterator) receiveWatchEvent(event *WatchEvent) {
	i.eventLock.Lock()
	defer i.eventLock.Unlock()

	i.events.Append(event)
}

func (i *Iterator) event() (event *WatchEvent, isDone bool, err error) {

	// check context cancel
	select {
	case _, isDone = <-i.done:
		return
	default:
	}

	// any pending event
	if event = i.eventNow(); event != nil {
		return
	}

	// wait for event, error or context cancel
	select {
	case _, isDone = <-i.done:
		return
	case <-i.isEvent.Ch():
		i.isEvent.Open()
	}
	err = i.error()
	event = i.eventNow()

	return
}

func (i *Iterator) eventNow() (event *WatchEvent) {
	i.eventLock.Lock()
	defer i.eventLock.Unlock()

	if len(i.events.Slice) == 0 {
		return
	}
	event = i.events.Slice[0]
	i.events.Slice[0] = nil
	i.events.Slice = i.events.Slice[1:]

	return
}

// errFn receives errors
func (i *Iterator) errFn(err error) {
	for {
		var errp = i.err.Load()
		var err1 error
		if errp == nil {
			err1 = perrors.AppendError(*errp, err)
		} else {
			err1 = err
		}
		if i.err.CompareAndSwap(errp, &err1) {
			break
		}
	}
	i.isEvent.Close()
}

func (i *Iterator) error() (err error) {
	if i.err.Load() != nil {
		if ep := i.err.Swap(nil); ep != nil {
			err = *ep
		}
	}
	return
}
