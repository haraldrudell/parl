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
	"github.com/haraldrudell/parl/ptime"
)

const (
	// shim instantiated but not watching or anything else
	shimIdle shimState = iota
	// shim is actively watching
	shimOpen
	// a thread had error, no watcher is active
	shimError
	// no watcher is active but Shutdown has not been invoked
	shimShuttingDown
	// Shutdown is about to end
	shimShutdownWinner
	// Shutdown has completed
	shimShutDown
)

// shimState trackes state of the shim
//   - shimIdle shimOpen shimError shimShuttingDown
//     shimShutdownWinner shimShutDown
type shimState uint32

// unique watcherShim ID 1…
var watchNo atomic.Uint64

// Watcher shim interfaces [fsnotify.Watcher]
//   - purpose is to completely encapsulate fsnotify implementation
//   - portable callback api
//   - errFn recieves: streamed fsnotify api errors, eventFunc errors, runtime errors and
//     fsnotify api close errors
//   - [WatcherShim.Watch] creates [fsnotify.Watcher].
//     Watch launches 2 threads listening to the api channels
//   - [WatcherShim.Add] watches individual paths.
//     A watched directory does not watch entries in its subdirectories
//   - [WatcherShim.List] returns absolute watched paths
//   - [Watcher.Shutdown] must be invoked
//   - Wat
type WatcherShim struct {
	// unique ID 1…
	ID uint64
	// invoked on every event
	//	- timestamp added
	//	- portable types
	eventFunc func(name string, op Op, t time.Time) (err error)
	// error sink
	errFn func(err error)

	// awaitable for running threads
	threadCounter sync.WaitGroup
	// closes on conclusion of Shutdown
	ShutdownComplete parl.Awaitable

	// watcherLock serializes open and close invocations
	watcherLock sync.Mutex
	// written behind watcherLock
	//	- except for shimShutdownWinner	shimShutDown
	state parl.Uint32[shimState]
	// written behind watcherLock
	watcher atomic.Pointer[fsnotify.Watcher]
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
	watcherShim.ID = watchNo.Add(1)
	watcherShim.eventFunc = eventFunc
	watcherShim.errFn = errFn
	watcherShim.ShutdownComplete = *parl.NewAwaitable()
	return
}

// Watch starts the Watcher. Thread-safe
//   - may only be invoked once
//   - on success, launches 2 goroutines and sets state to shimOpen
func (w *WatcherShim) Watch() (err error) {

	// fast outside lock check
	// create fsnotify watcher behind lock
	if err = w.shimOpenState(); err != nil {
		return // bad state return
	} else if err = w.open(); err != nil {
		return // already shutdown or failure return
	}

	return // watcher created return
}

// Add adds a path for watching
//   - if anything amiss, error
func (w *WatcherShim) Add(name string) (err error) {
	var watcher = w.watcher.Load()
	if w.state.Load() != shimOpen || watcher == nil {
		err = perrors.NewPF("shim idle, error or shutdown")
		return
	}
	if err = watcher.Add(name); perrors.IsPF(&err, "fsmotify.Watcher.Add %w", err) {
		return
	}

	return
}

// List returns the currently watched paths. Thread-Safe
//   - if watcher not open or errored or shutdown: nil
func (w *WatcherShim) List() (paths []string) {
	var watcher = w.watcher.Load()
	if watcher == nil {
		return // shim not open
	} else if w.state.Load() != shimOpen {
		return // shim in error or shutdown state
	}
	paths = watcher.WatchList()

	return
}

// Shutdown closes the watcher
//   - does not return prior to watcher closed
//   - thread-safe
func (w *WatcherShim) Shutdown() {

	var state = w.state.Load()
	switch state {
	case shimShutDown:
		return // shutdown already complete
	case shimIdle, shimOpen:
		// shim needs a close
		//	- state goes to shimShuttingDown
		w.close()
	}

	// many threads may arrive here
	//	- fsnotify-watcher is closed
	//	- need to wait for threads to exit
	for {
		if state = w.state.Load(); state == shimShutDown {
			return // other thread already completed
		} else if state == shimShutdownWinner {
			// another thread won. Wait for it to complete
			<-w.ShutdownComplete.Ch()
			return
		} else if w.state.CompareAndSwap(state, shimShutdownWinner) {
			break // this thread won
		}
	}

	// wait for threads to exit, then trigger awaitable
	w.threadCounter.Wait()
	w.state.Store(shimShutDown)
	w.ShutdownComplete.Close()
}

// errorThread reads the error channel until end
func (w *WatcherShim) errorThread(errCh <-chan error) {
	defer w.endingThread()
	// infallible because likely, errFn paniced
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, parl.Infallible)

	for err := range errCh {
		w.errFn(err)
	}
}

// event thread reads the event channel until end
//   - provides timestamped values
func (w *WatcherShim) eventThread(eventCh <-chan fsnotify.Event) {
	defer w.endingThread()
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, w.errFn)

	// wait for event
	//	- received as value, a string and an int
	for fsnotifyEvent := range eventCh {
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

// shimOpenState ensures that shim is idle
func (w *WatcherShim) shimOpenState() (err error) {
	switch w.state.Load() {
	case shimIdle:
	case shimOpen:
		return perrors.NewPF("invoked more than once")
	case shimError:
		return perrors.NewPF("had error")
	default:
		return perrors.NewPF("after shutdown")
	}
	return
}

// endingThread initiates close
//   - may set shimError state if not already shutdown
//   - close will end both threads
func (w *WatcherShim) endingThread() {
	w.threadCounter.Done()

	// fast outside lock check
	switch w.state.Load() {
	case shimIdle, shimOpen:
	default:
		return // shim already errored or shutdown
	}

	// this thread will close
	w.close(shimError)
}

// open creates fsnotify.Watcher
//   - good return is shimOpen state
func (w *WatcherShim) open() (err error) {
	w.watcherLock.Lock()
	defer w.watcherLock.Unlock()

	if err = w.shimOpenState(); err != nil {
		return
	}
	var watcher *fsnotify.Watcher
	if watcher, err = fsnotify.NewWatcher(); perrors.IsPF(&err, "fsnotify.NewWatcher %w", err) {
		return
	}
	w.watcher.Store(watcher)
	// launch goroutines
	w.threadCounter.Add(2)
	go w.errorThread(watcher.Errors)
	go w.eventThread(watcher.Events)
	w.state.Store(shimOpen)

	return
}

// close closes the watcher
//   - idempotent, thread-safe
//   - sets state shimShuttingDown or state0 if present/shimError
//   - [fsnotify.Watcher.Close] is idempotent
func (w *WatcherShim) close(state0 ...shimState) {
	var nextState shimState
	if len(state0) > 0 && state0[0] == shimError {
		nextState = shimError
	} else {
		nextState = shimShuttingDown
	}
	w.watcherLock.Lock()
	defer w.watcherLock.Unlock()

	// inside-lock state check
	var state = w.state.Load()
	switch state {
	case shimIdle, shimOpen: // these need close
	case shimError:
		if nextState == shimError {
			return // was another error, already in error state
		}
		// already closed by error: set shutdown
		w.state.Store(shimShuttingDown)
		return
	default:
		return // some other thread already closed
	}
	// this thread is selected to close

	// is there a watcher to close?
	var watcher = w.watcher.Load()
	if watcher == nil {
		w.state.Store(nextState)
		return // Watch was never invoked
	}

	// close the watcher
	var err = watcher.Close()
	w.state.Store(nextState)
	if err != nil {
		w.errFn(err)
	}
}
