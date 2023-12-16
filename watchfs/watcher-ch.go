/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"regexp"
	"sync"
)

type WatcherCh struct {
	ch        chan any
	watcher   *Watcher
	eventLock sync.Mutex
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
	return UnpackEvents(<-w.ch)
}

// eventFunc sends on w.ch
func (w *WatcherCh) eventFunc(event *WatchEvent) {
	var anyEvent any
	w.eventLock.Lock()
	defer w.eventLock.Unlock()

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

func (w *WatcherCh) Shutdown() { w.watcher.Shutdown() }
