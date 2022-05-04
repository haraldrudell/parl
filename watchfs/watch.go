/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/pstrings"
	"github.com/haraldrudell/parl/ptime"
)

type Watch struct {
	Now          time.Time
	ID           int64
	dir0         string
	cleanDir     string
	abs          string
	events       chan *WatchEvent
	errChan      chan error
	watcher      fsnotify.Watcher
	filter       fsnotify.Op
	shutCh       chan struct{}
	ctx          context.Context
	shutdownLock sync.Once
	wgThread     sync.WaitGroup
	isShutdown   parl.AtomicBool
}

var watchNo int64

/*
NewWatch produces events from a file-system entry and its child directories.
WatchEvents must be listened for on the .Events() channel until it closes.
Events that can be listened for are WatchOpAll (default: 0) Create Write Remove Rename Chmod.
errors must be listened for on the .Errors() channel until it closes.
Close the watcher by canceling the context or invoking .Shutdown().

220315 github.com/fsnotify/fsnotify v1.4.9 does not support macOS
220315 use the old github.com/fsnotify/fsevents v0.1.1
*/
func NewWatch(directory string, filter fsnotify.Op, ctx context.Context) (watch *Watch) {

	// listen to file system
	if ctx == nil {
		ctx = context.Background()
	}
	wa := Watch{
		Now:      time.Now(),
		ID:       atomic.AddInt64(&watchNo, 1),
		dir0:     directory,
		cleanDir: filepath.Clean(directory),
		events:   make(chan *WatchEvent),
		errChan:  make(chan error),
		filter:   filter,
		shutCh:   make(chan struct{}),
		ctx:      ctx,
	}
	if abs, err := filepath.Abs(directory); err != nil {
		panic(parl.Errorf("filepath.Abs: %q '%w'", directory, err))
	} else {
		wa.abs = abs
	}
	if err := wa.startWatcher(); err != nil {
		panic(err) // failure setting up directory three watchers
	}
	parl.Info("%s WatchFS directories added\n", ptime.NsLocal(wa.Now))

	// strat thread emitting event and errors
	wa.wgThread.Add(1)
	go wa.fsnotifyListener()

	return &wa
}

// Events returns a receiver channel for pfs.WatchEvent
func (wa *Watch) Events() (events <-chan *WatchEvent) {
	return wa.events
}

func (wa *Watch) Errors() (errCh <-chan error) {
	return wa.errChan
}

func (wa *Watch) Shutdown() {
	wa.shutdownLock.Do(wa.shutdown)
}

func (wa *Watch) startWatcher() (err error) {
	var watcher *fsnotify.Watcher
	defer func() {
		if err != nil && watcher != nil {
			if e := wa.watcher.Close(); e != nil {
				err = perrors.AppendError(err, parl.Errorf("watcher.Close: %w", e))
			}
		}
	}()
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return parl.Errorf("fsnotify.NewWatcher: '%w'", err)
	}
	wa.watcher = *watcher

	return wa.scan()
}

func (wa *Watch) scan() (err error) {
	// scan Dir for all subdirectories
	// neither Linux and macOS can watch a directory tree so
	// each watched directory needs to be referenced
	var dirs []string
	if dirs, err = pfs.Dirs(wa.dir0); err != nil {
		return
	}
	parl.Info("watching: %s\n", pstrings.QuoteList(dirs))

	// add directories
	for _, dir := range dirs {
		if err = wa.watcher.Add(dir); err != nil {
			err = parl.Errorf("setWatch Linux watcher.Add: '%w'", err)
			return
		}
	}
	return
}

func (wa *Watch) fsnotifyListener() {
	var err error
	defer func() {
		wa.wgThread.Done() // this releases .Shutdown()
		wa.Shutdown()      // invoke Shutdown if that has not already happened
		if e := parl.HandlePanic(func() { close(wa.errChan) }); e != nil {
			parl.Log("close err chan: %s", e.Error())
		}
	}()
	defer parl.Recover2(parl.Annotation(), &err, func(e error) { wa.errChan <- e })
	defer func() {
		close(wa.events)
	}()
	defer func() {
		if e := wa.watcher.Close(); e != nil {
			wa.errChan <- parl.Errorf("fsnotify.Watcher.Close: '%w'", e)
		}
	}()

	for {
		select {
		case inEvent, ok := <-wa.watcher.Events:
			now := time.Now()
			if !ok {
				if !wa.isShutdown.IsTrue() {
					wa.errChan <- parl.Errorf("%d event channel closed", wa.ID)
				}
				break
			}
			parl.Info("%s %s\n", ptime.NsLocal(now), inEvent)

			// filter
			if wa.filter != WatchOpAll && inEvent.Op&wa.filter == 0 {
				continue
			}

			// emit
			wa.events <- &WatchEvent{
				At:        now,
				ID:        uuid.New(),
				BaseName:  filepath.Base(inEvent.Name),
				AbsName:   filepath.Join(wa.abs, inEvent.Name),
				CleanName: filepath.Join(wa.cleanDir, inEvent.Name),
				Op:        Op(inEvent.Op).String(),
				//Event:     inEvent,
			}
			continue
		case <-wa.ctx.Done(): // shutdown via context
		case <-wa.shutCh: // shutdown via .Shutdown()
			if !wa.isShutdown.IsTrue() {
				wa.errChan <- parl.New("shutdown channel unexpectedly closed")
			}
		} // select
		break
	} // for
}

func (wa *Watch) shutdown() {
	wa.isShutdown.Set()
	close(wa.shutCh)
	wa.wgThread.Wait()
}
