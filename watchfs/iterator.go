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

// Iterator is a file-system watcher with Go for-statement iterative api
type Iterator struct {
	// path is the file-system entry watched
	//	- may be relative, unclean and contain symlinks
	path string
	// watcher is a file-system watcher with callback api
	watcher Watcher
	// if Watch has been invoked
	isWatching atomic.Bool
	// triggers on event receive
	hasEvent parl.CyclicAwaitable
	// err receives errors from the watcher api in real-time
	err atomic.Pointer[error]
	// true if asyncCancel was invoked
	isAsyncCancel parl.Awaitable
	// closes on context cancel
	ctx context.Context

	// eventLock makes events thread-safe
	eventLock sync.Mutex
	// events buffers wacther events that may arrive many at a time
	events pslices.Shifter[*WatchEvent]
}

// NewIterator returns a file-system watch-event iterator
//   - blocking iterator for use with go for-statement
//   - optional context for cancelation
//   - filter [WatchOpAll] (default: 0) is: Create Write Remove Rename Chmod
//     it can also be a bit-coded value.
//   - ignores is a regexp for the absolute filename.
//     It is applied while scanning directories, not for individual events.
//   - thread-safe
//   - consumers are expected to use:
//   - — [NewIterator] using a Go for-statement iterative api
//   - — [NewWatcherCh] using Go channel api
//   - — [NewWatcher] using callback api
//
// Usage:
//
//	var iterator = watchfs.NewIterator(somePath, watchfs.WatchOpAll, watchfs.NoIgnores, ctx)
//	defer iterator.Cancel(&err)
//	for watchEvent, _ := iterator.Init(); iterator.Cond(&watchEvent); {…
func NewIterator(path string, filter Op, ignores *regexp.Regexp, ctx ...context.Context) (iterator iters.Iterator[*WatchEvent]) {

	i := Iterator{
		path: path,
	}
	if len(ctx) > 0 {
		i.ctx = ctx[0]
	}
	NewWatcher(
		filter, ignores, i.receiveEventFromWatcher, newErrorSink(&i),
		&i.watcher,
	)
	return iters.NewFunctionIterator(i.iteratorFunction, i.asyncCancel)
}

// iteratorFunction awaits the next event
//   - iteratorFunction invocations are serialized by the function iterator.
//     Therefore, iteratorFunction is a critical section with only one thread
//     executing at any time
//   - any thread may invoke iteratorFunction why it
//     must be thread-safe
//   - iteratorFunction is blocking but can be aborted by
//     asyncCancel
//   - thread-safe
func (i *Iterator) iteratorFunction(isCancel bool) (fsEvent *WatchEvent, err error) {

	//handle cancel and error
	if isCancel {
		parl.Debug("received isCancel true: watcher.Shutdown")
		i.watcher.Shutdown()
	}
	err = i.error() // collect any Shutdown errors
	if isCancel || i.isAsyncCancel.IsClosed() || err != nil {
		if !isCancel && err == nil {
			err = parl.ErrEndCallbacks
		}
		return // cancel or error return: will not be invoked again
	}

	// ensure Watching
	if !i.isWatching.Load() {
		i.isWatching.Store(true)
		parl.Debug("invoking watcher.Watch %q", i.path)
		if err = i.watcher.Watch(i.path); err != nil {
			i.watcher.Shutdown()
			// ignore i.error, an error is already present
			return // bad watch return
		}
	}

	// get any queued up event
	if fsEvent = i.getEvent(); fsEvent != nil {
		return // success return
	}
	// arm cyclable
	i.hasEvent.Open()
	// make sure any event is collected prior to arming recyclable
	if fsEvent = i.getEvent(); fsEvent != nil {
		return // success return
	}

	// wait for next event
	var maybeErr error
	var done <-chan struct{}
	if i.ctx != nil {
		done = i.ctx.Done()
	}
	select {
	case <-i.hasEvent.Ch():
		if fsEvent = i.getEvent(); fsEvent != nil {
			return // success return
		}
		maybeErr = perrors.NewPF("Bad state")
	case <-done:
		maybeErr = i.ctx.Err()
	case <-i.isAsyncCancel.Ch():
		maybeErr = parl.ErrEndCallbacks
	}
	i.watcher.Shutdown()
	if err = i.error(); err == nil {
		err = maybeErr
	}

	return // cancel error return
}

// receiveWatchEvent receives events from the watcher
//   - thread-safe
func (i *Iterator) receiveEventFromWatcher(event *WatchEvent) {
	i.eventLock.Lock()
	defer i.eventLock.Unlock()

	i.events.Append(event)
	i.hasEvent.Close()
}

// getEvent gets an event from the buffer
func (i *Iterator) getEvent() (event *WatchEvent) {
	i.eventLock.Lock()
	defer i.eventLock.Unlock()

	if len(i.events.Slice) == 0 {
		return // no event return
	}
	event = i.events.Slice[0]
	i.events.Slice[0] = nil
	i.events.Slice = i.events.Slice[1:]

	return // has event return
}

// errFn receives errors from the watcher api
//   - invoked at any time
//   - thread-safe
func (i *Iterator) addError(err error) {
	// append the error to i.err
	for {
		var errp = i.err.Load()
		var err1 error
		if errp != nil {
			err1 = perrors.AppendError(*errp, err)
		} else {
			err1 = err
		}
		if i.err.CompareAndSwap(errp, &err1) {
			break // successfully updated
		}
	}
	i.asyncCancel()
}

// error retrives any errors from watcher api
func (i *Iterator) error() (err error) {
	if ep := i.err.Load(); ep != nil {
		err = *ep
	}
	return
}

// asyncCancel is used to note cancel since watcher is blocking
//   - can be invoked at any time
//   - thread-safe
func (i *Iterator) asyncCancel() {
	i.isAsyncCancel.Close()
	i.watcher.Shutdown()
}
