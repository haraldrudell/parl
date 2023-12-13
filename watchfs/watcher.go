/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package watchfs provides a file-system watcher for Linux and macOS.
package watchfs

import (
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

var watchNo int64 // atomic

type Watcher struct {
	abs           string
	cleanDir      string
	filter        Op
	ignores       *regexp.Regexp
	errFn         func(err error)
	eventFn       func(event *WatchEvent)
	eventFn0      func(event *WatchEvent)
	watcher       *fsnotify.Watcher
	ID            int64
	watcherClosed atomic.Bool
	wg            sync.WaitGroup
}

// NewWatcher0 returns a file system watcher.
// NewWatcher0 does not implicitly watch anything.
// errFn must be thread-safe.
// eventFn must be thread-safe.
func NewWatcher0(
	ignores *regexp.Regexp,
	eventFn func(event *WatchEvent),
	errFn func(err error)) (watch *Watcher) {

	w := Watcher{
		errFn:   errFn,
		eventFn: eventFn,
		ID:      atomic.AddInt64(&watchNo, 1),
	}
	var err error
	if w.watcher, err = fsnotify.NewWatcher(); err != nil {
		panic(perrors.Errorf("fsnotify.NewWatcher: '%w'", err))
	}
	w.wg.Add(2)
	go w.errorThread()
	go w.eventThread()
	return &w
}

func (w *Watcher) List() (paths []string) {
	return w.watcher.WatchList()
}

func (w *Watcher) Shutdown() {
	if w.watcherClosed.CompareAndSwap(false, true) {
		var err error
		if err = w.watcher.Close(); err != nil {
			w.errFn(perrors.Errorf("watcher.Close: %w", err))
		}
	}
	w.wg.Wait()
}

func (w *Watcher) errorThread() {
	w.wg.Done()
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, parl.Infallible)

	errCh := w.watcher.Errors
	for {
		if err, ok := <-errCh; !ok {
			return
		} else {
			w.errFn(err)
		}
	}
}
func (w *Watcher) eventThread() {
	w.wg.Done()
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, w.errFn)

	events := w.watcher.Events
	for {

		// wait for event
		fsnotifyEvent, ok := <-events
		if !ok {
			return // events channel closed exit
		}

		// debug print
		now := time.Now()
		parl.Debug("%s %s", ptime.NsLocal(now), fsnotifyEvent)

		// ignore filter
		if w.ignores != nil && w.ignores.MatchString(fsnotifyEvent.Name) {
			continue // ignore pattern skip
		}

		// send event
		w.sendEvent(NewWatchEvent(&fsnotifyEvent, now, w))
	}
}

func (w *Watcher) sendEvent(ev *WatchEvent) {
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, w.errFn)

	w.eventFn(ev)
}
