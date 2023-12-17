/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"regexp"
	"sync"
	"sync/atomic"
)

// WatcherCh provides a channel api for a file-system watcher
type WatcherCh struct {
	// ch sends *WatchEvent or []*WatchEvent values
	ch chan any
	// watcher is a file-system  watcher with portable callback api
	watcher *Watcher

	// eventLock ensures serialization of read-write-back to ch
	// and close
	eventLock sync.Mutex
	// if ch is closed
	//	- written behind eventLock
	isClosed atomic.Bool
}

// NewWatcherCh returns a file-system watcher using channel mechanic
//   - filter [WatchOpAll] (default: 0) is: Create Write Remove Rename Chmod.
//     it can also be a bit-coded value.
//   - ignores is a regexp for the absolute filename.
//     it is applied while scanning directories.
//   - errFn must be thread-safe.
//   - stores self-referencing pointers
func NewWatcherCh(
	filter Op, ignores *regexp.Regexp,
	errFn func(err error),
) (watcherCh *WatcherCh) {
	w := WatcherCh{ch: make(chan any, 1)}
	w.watcher = NewWatcher(filter, ignores, w.eventFunc, errFn)
	return &w
}

func (w *WatcherCh) Watch(path string) (err error) { return w.watcher.Watch(path) }

// Ch returns either *WatchEvent or []*WatchEvent
//   - Thread-safe
func (w *WatcherCh) Ch() (ch <-chan any) { return w.ch }

// Get receives one or more watch events
//   - blocks until an event is available or the watcher ends.
//     For non-blocking designs, use [WatcherCh.Ch]
//   - if watchEvent non-nil, it is the single event returned
//   - if len(watchEvents) > 0, it holds 2 or more returned events
//   - if watchEvent and watchEvents are both nil, the watcher has ended
//   - Thread-safe
func (w *WatcherCh) Get() (watchEvent *WatchEvent, watchEvents []*WatchEvent) {
	var any, ok = <-w.ch
	if !ok {
		return
	}
	return UnpackEvents(any)
}

func (w *WatcherCh) Shutdown() {
	// shutdown callback api
	w.watcher.Shutdown()

	// ensure channel is closed
	if w.isClosed.Load() {
		return
	}
	w.eventLock.Lock()
	defer w.eventLock.Unlock()

	if w.isClosed.Load() {
		return
	}

	close(w.ch)
	w.isClosed.Store(true)
}

// eventFunc sends on w.ch
func (w *WatcherCh) eventFunc(event *WatchEvent) {
	var anyEvent any
	w.eventLock.Lock()
	defer w.eventLock.Unlock()

	if w.isClosed.Load() {
		return
	}
	select {
	case anyEvent = <-w.ch:
		// if single event, create slice
		if watchEvent, ok := anyEvent.(*WatchEvent); ok {
			w.ch <- []*WatchEvent{watchEvent, event}
			return
		}
		// already multiple events, append to slice
		w.ch <- append(anyEvent.([]*WatchEvent), event)
	default:
		w.ch <- event
	}
}
