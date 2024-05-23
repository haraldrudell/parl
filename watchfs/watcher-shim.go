/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package watchfs provides a file-system watcher for Linux and macOS.
package watchfs

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/ptime"
)

const (
	// number of threads run by shim
	threadCount = 2
)

// unique watcherShim ID 1…
var watchNo atomic.Uint64

// Watcher shim interfaces [fsnotify.Watcher]
//   - purpose is to completely encapsulate fsnotify implementation
//   - portable callback api
//   - errFn receives:
//   - — streamed fsnotify api errors
//   - — eventFunc callback errors
//   - — runtime errors and
//   - — fsnotify api close errors
//   - [WatcherShim.Watch] launches 2 threads listening to the api channels
//   - [WatcherShim.Add] watches individual paths.
//     A watched directory does not watch entries in its subdirectories
//   - [Watcher.Shutdown] must be invoked after successful Watch
//     to release resources.
//     Shutdown does not return until resources are released
type WatcherShim struct {
	// unique ID 1…
	ID uint64
	// invoked on every event
	//	- timestamp added
	//	- portable types
	eventFunc func(name string, op Op, t time.Time) (err error)
	// error sink
	errorSink parl.ErrorSink1
	// ensures Watch only invoked once
	isWatchInvoked atomic.Bool
	// awaitable for running threads
	threadCounter sync.WaitGroup
	// selects winner thread for shutdown
	isShutdownInvoked atomic.Bool
	// awaitable that triggers on conclusion of Shutdown
	shutdownComplete parl.Awaitable
	watcher          atomic.Pointer[fsnotify.Watcher]
}

// NewWatcherShim returns a file system watcher using portable types
//   - NewWatcherShim is an abstraction hiding the file-system watch implementation.
//     Consumers are expected to use:
//   - — [NewIterator] using a Go for-statement iterative api
//   - — [NewWatcherCh] using Go channel api
//   - — [NewWatcher] using callback api
//   - on any error, the watcherShim shuts down
//   - [WatcherShim.Watch] starts watching
//   - [WatcherShim.Add] watches a a path
//   - [WatcherShim.List] lists watched paths
//   - [WatcherShim.Shutdown] must be invoked to release resources
func NewWatcherShim(
	fieldp *WatcherShim,
	eventFunc func(name string, op Op, t time.Time) (err error),
	errorSink parl.ErrorSink1,
) (watcherShim *WatcherShim) {
	if eventFunc == nil {
		panic(parl.NilError("eventFunc"))
	} else if errorSink == nil {
		panic(parl.NilError("errFn"))
	}
	if fieldp != nil {
		watcherShim = fieldp
		*watcherShim = WatcherShim{}
	} else {
		watcherShim = &WatcherShim{}
	}
	watcherShim.ID = watchNo.Add(1)
	watcherShim.eventFunc = eventFunc
	watcherShim.errorSink = errorSink
	return
}

// Watch starts the Watcher. Thread-safe
//   - may only be invoked once
//   - [fsnotify.NewWatcher] launches goroutines
//   - on success, launches 2 goroutines and sets state to shimOpen
func (w *WatcherShim) Watch() (err error) {

	// ensure only invoked once
	if !w.isWatchInvoked.CompareAndSwap(false, true) {
		err = perrors.NewPF("invoked more than once")
		return // invoked more than once return
	}

	// use threadCounter as mechanic to make shutdown wait
	var didLaunchThreads bool
	w.threadCounter.Add(threadCount)
	defer w.watchEnd(&didLaunchThreads)

	// check for shutdown
	if w.isShutdownInvoked.Load() {
		err = perrors.NewPF("shim shut down")
		return // shutdown in progress or complete return
	}

	// create watcher
	var watcher *fsnotify.Watcher
	if watcher, err = fsnotify.NewWatcher(); perrors.IsPF(&err, "fsnotify.NewWatcher %w", err) {
		return // new watcher failed return
	}
	w.watcher.Store(watcher)

	// launch goroutines
	go w.errorThread(watcher.Errors)
	go w.eventThread(watcher.Events)
	didLaunchThreads = true

	return // watcher created return
}

// Add adds a path for watching
//   - if anything amiss, error
func (w *WatcherShim) Add(name string) (err error) {

	// check state
	var watcher = w.watcher.Load()
	if watcher == nil || w.isShutdownInvoked.Load() {
		err = perrors.NewPF("shim idle, error or shutdown")
		return // unable return
	}

	// fsnotify.List does not clean or abs the result
	//	- do that here
	//	- bad symlinks or watching what does not exist will
	//		fail prematurely
	var absEval string
	if absEval, err = pfs.AbsEval(name); err != nil {
		return
	}

	// watch the path
	if err = watcher.Add(absEval); perrors.IsPF(&err, "fsnotify.Watcher.Add %w", err) {
		return // Add failure return
	}

	return // success return
}

// List returns the currently watched paths. Thread-Safe
//   - if watcher not open or errored or shutdown: nil
func (w *WatcherShim) List() (paths []string) {

	// check state
	var watcher = w.watcher.Load()
	if watcher == nil || w.isShutdownInvoked.Load() {
		return // unable return
	}

	// best effort WatchList
	paths = watcher.WatchList()

	return // success return
}

// Shutdown closes the watcher
//   - does not return prior to watcher closed and resources released
//   - thread-safe
func (w *WatcherShim) Shutdown() {

	// check shutdown state
	if w.shutdownComplete.IsClosed() {
		return // already shutdown return
	} else if !w.isShutdownInvoked.CompareAndSwap(false, true) {
		// await shutdown completion by oher thread
		<-w.shutdownComplete.Ch()
		return
	}
	// this thread won shutdown race
	defer w.shutdownComplete.Close()

	// closing the watcher will cause threads to exit
	w.ensureWatcherClose()

	// wait for threads to exit, then trigger awaitable
	w.threadCounter.Wait()
}

// watchEnd releases threadCounter if threads were not launched
func (w *WatcherShim) watchEnd(didLaunchThreads *bool) {
	if *didLaunchThreads {
		return
	}
	w.threadCounter.Add(-threadCount)
}

// errorThread reads the error channel until end
//   - runtime errors are printed to standard error
func (w *WatcherShim) errorThread(errCh <-chan error) {
	defer w.endingThread()
	// infallible because likely, errFn paniced
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, parl.Infallible)

	for err := range errCh {
		w.errorSink.AddError(err)
	}
}

// event thread reads the event channel until end
//   - provides timestamped values
func (w *WatcherShim) eventThread(eventCh <-chan fsnotify.Event) {
	defer w.endingThread()
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, w.errorSink)

	// wait for event
	//	- received as value, a string and an int
	for fsnotifyEvent := range eventCh {
		var now = time.Now()

		// debug print
		parl.OnDebug(func() string { return parl.Sprintf("%s %s", ptime.NsLocal(now), fsnotifyEvent) })

		// send event
		if err := w.eventFunc(fsnotifyEvent.Name, Op(fsnotifyEvent.Op), now); err != nil {
			w.errorSink.AddError(perrors.ErrorfPF("eventFunc %w", err))
			return
		}
	}
}

// endingThread initiates close
//   - may set shimError state if not already shutdown
//   - close will end both threads
func (w *WatcherShim) endingThread() {
	w.threadCounter.Done()

	// closing the watcher will exit all threads
	w.ensureWatcherClose()
}

// ensureWatcherClose closes the watcher
//   - idempotent, thread-safe
//   - upon return, w.watcher is nil.
//     The watcher itself may not have closed yet
func (w *WatcherShim) ensureWatcherClose() {

	// select winner thread to close
	var watcher = w.watcher.Load()
	if watcher == nil {
		return // watcher is not created, being closed or closed
	} else if !w.watcher.CompareAndSwap(watcher, nil) {
		return // watcher is closed or being closed
	}
	// this thread is selected to close

	// close the watcher
	var err = watcher.Close()
	if err != nil {
		w.errorSink.AddError(err)
	}
}
