/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package watchfs provides a file-system watcher for Linux and macOS.
package watchfs

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

var watchNo uint64 // atomic

// Watcher shim interfaces [fsnotify.Watcher]
type WatcherShim struct {
	ID        uint64
	Awaitable parl.Awaitable
	eventFunc func(name string, op Op, t time.Time) (err error)
	errFn     func(err error)
	watcher   atomic.Pointer[fsnotify.Watcher]
	counter   atomic.Uint32
}

// NewWatcherShim returns a file system watcher using portable types
//   - eventFunc receives timestamped events in real time and must be thread-safe
//   - errFn receives any errors in real time and must be thread-safe
//   - on any error, the watcherShim shuts down
//   - use [NewWatcherCh] or for callback [NewWatcher].
//     NewWatcherShim is an abstraction hiding the implementation.
func NewWatcherShim(
	fieldp *WatcherShim,
	eventFunc func(name string, op Op, t time.Time) (err error),
	errFn func(err error),
) (watcherShim *WatcherShim) {
	if eventFunc == nil {
		panic(parl.NilError("eventFunc"))
	} else if errFn == nil {
		panic(parl.NilError("errFn"))
	}
	if fieldp != nil {
		watcherShim = fieldp
		*watcherShim = WatcherShim{}
	} else {
		watcherShim = &WatcherShim{}
	}
	watcherShim.errFn = errFn
	watcherShim.eventFunc = eventFunc
	watcherShim.ID = atomic.AddUint64(&watchNo, 1)
	watcherShim.Awaitable = *parl.NewAwaitable()
	return
}

// Watch starts the Watcher. Thread-safe
func (w *WatcherShim) Watch() (err error) {
	if w.watcher.Load() != nil {
		panic(perrors.NewPF("invoked more than once"))
	}
	var watcher *fsnotify.Watcher
	if watcher, err = fsnotify.NewWatcher(); perrors.IsPF(&err, "fsnotify.NewWatcher %w", err) {
		return
	}
	if !w.watcher.CompareAndSwap(nil, watcher) {
		panic(perrors.NewPF("invoked more than once"))
	}
	w.counter.Add(2)
	go w.errorThread()
	go w.eventThread()

	return
}

func (w *WatcherShim) Add(name string) (err error) {
	var watcher = w.watcher.Load()
	if watcher == nil {
		return
	}
	if err = watcher.Add(name); perrors.IsPF(&err, "fsmotify.Watcher.Add %w", err) {
		return
	}

	return
}

// List returns the currently watched paths. Thread-Safe
func (w *WatcherShim) List() (paths []string) {
	var watcher = w.watcher.Load()
	if watcher == nil {
		return
	}
	paths = watcher.WatchList()

	return
}

func (w *WatcherShim) Shutdown() {
	if w.Awaitable.IsClosed() {
		return // already closed
	}
	w.counter.Add(1)
	w.close()
}

// errorThread reads the error channel until end
func (w *WatcherShim) errorThread() {
	defer w.close()
	// infallible because likely, errFn paniced
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, parl.Infallible)

	for err := range w.watcher.Load().Errors {
		w.errFn(err)
	}
}

// event thread reads the event channel until end
//   - provides timestamped values
func (w *WatcherShim) eventThread() {
	defer w.close()
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, w.errFn)

	// wait for event
	//	- received as value, a string and an int
	for fsnotifyEvent := range w.watcher.Load().Events {
		var now = time.Now()

		// debug print
		parl.OnDebug(func() string { return parl.Sprintf("%s %s", ptime.NsLocal(now), fsnotifyEvent) })

		// send event
		if err := w.eventFunc(fsnotifyEvent.Name, Op(fsnotifyEvent.Op), now); err != nil {
			w.errFn(perrors.ErrorfPF("eventFunc %w", err))
			return
		}
	}
}

// close closes the watcher.
//   - [fsnotify.Watcher.Close] is idempotent
func (w *WatcherShim) close() {
	var err error
	if parl.Close(w.watcher.Load(), &err); perrors.IsPF(&err, "fsnotify Close %w", err) {
		w.errFn(err)
	}
	if w.counter.Add(math.MaxUint32) == 0 {
		w.Awaitable.Close()
	}
}
