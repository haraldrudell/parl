/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/ptime"
)

var NoIgnores *regexp.Regexp

// 220505 github.com/fsnotify/fsnotify v1.5.4
// 220315 github.com/fsnotify/fsnotify v1.4.9 does not support macOS
// 220315 use the old github.com/fsnotify/fsevents v0.1.1

// Watcher implements a file-system watcher using callback api
//   - consumers are expected to use:
//   - — [NewIterator] using a Go for-statement iterative api
//   - — [NewWatcherCh] using Go channel api
//   - — [NewWatcher] using callback api
//   - eventFn receives filtered events and must be thread-safe
//   - if directories are created, Watcher adds those to being watched
//   - —
//   - the fsnotify api provides two channels so threading is required
//   - the Errors channel is unbuffered
//   - the Events channel is typically unbuffered but can be buffered
//   - Events are sent by value with name and Op only
//   - — no timestamp
//   - — no unique identifier
//   - watchers are not recursive into subdirectories
//   - — to detect new entires, all child-directories must be watched
type Watcher struct {
	eventFn   func(event *WatchEvent)
	errorSink parl.ErrorSink1
	ignores   *regexp.Regexp
	filter    Op

	// addLock serializes Watch-create and Shutdown
	addLock    sync.Mutex
	isShutdown bool
	WatcherShim
}

// NewWatcher provides a channel sending file-system events from a file-system entry and its child directories.
//   - consider using [Iterator] or [NewWatcherCh] for callback-free architecture
//   - filter [WatchOpAll] (default: 0) is: Create Write Remove Rename Chmod.
//     it can also be a bit-coded value.
//   - ignores is a regexp for the absolute filename.
//     it is applied while scanning directories.
//   - eventFn must be thread-safe.
//     this means that any storing and subsequent retrieval of event
//     must be thread-safe, protected by go, Mutex or atomic
//   - errFn must be thread-safe.
//   - Close the watcher by canceling the context or invoking .Shutdown().
//     This means that any storing and subequent retrieval of the err value
//     must be thread-safe, protected by go, Mutex or atomic
func NewWatcher(
	filter Op, ignores *regexp.Regexp,
	eventFn func(event *WatchEvent), errorSink parl.ErrorSink1,
) (watcher *Watcher) {
	return &Watcher{
		eventFn:   eventFn,
		errorSink: errorSink,
		filter:    filter,
		ignores:   ignores,
	}
}

// Watch adds file-system entry-watchers
//   - entry is the file-system location being watched, absolute or relative.
//     If a directory, all subdirectories are watched, too.
func (w *Watcher) Watch(entry string) (err error) {
	defer w.shutdownOnErr(&err)

	// initialize api if not already done
	if w.WatcherShim.ID == 0 {
		// check that filter is valid
		if _, err = w.filter.fsnotifyOp(); err != nil {
			return
		}
		if err = w.create(); err != nil {
			return
		}
	}

	// get entry following symlinks
	var fsFileInfo fs.FileInfo
	if fsFileInfo, err = os.Stat(entry); perrors.IsPF(&err, "os.Stat %w", err) {
		return
	}

	// if not a directory, watch it
	if !fsFileInfo.IsDir() {
		if err = w.WatcherShim.Add(entry); err != nil {
			return
		}
		return // watching non-directory return
	}

	// scan for directories
	// neither Linux and macOS can watch a directory tree so
	// each watched directory needs to be referenced
	var now = time.Now()
	var iterator = pfs.NewDirIterator(entry)
	defer iterator.Cancel(&err)
	for resultEntry, _ := iterator.Init(); iterator.Cond(&resultEntry); {

		// scan directory for all subdirectories
		if w.ignores != nil && w.ignores.MatchString(resultEntry.Abs) {
			continue
		}
		if err = w.WatcherShim.Add(resultEntry.Abs); err != nil {
			return
		}
	}

	var t1 = time.Now()
	var d = t1.Sub(now).Round(100 * time.Millisecond)
	parl.Debug("%s WatchFS directories added in %s\n", ptime.NsLocal(now), d)

	return
}

func (w *Watcher) Shutdown() {
	w.addLock.Lock()
	defer w.addLock.Unlock()

	w.isShutdown = true
	if w.WatcherShim.ID != 0 {
		w.WatcherShim.Shutdown()
	}
}

func (w *Watcher) create() (err error) {
	w.addLock.Lock()
	defer w.addLock.Unlock()

	if w.WatcherShim.ID != 0 || w.isShutdown {
		return
	}
	// invoke Watch
	err = NewWatcherShim(&w.WatcherShim, w.filterEvent, w.errorSink).Watch()
	return
}

func (w *Watcher) filterEvent(name string, op Op, t time.Time) (err error) {

	// if it is CREATE of a directory, add that directory to be watched, too
	if op&Create != 0 {
		if err = w.addCreatedDirectoriesToWatcher(name); err != nil {
			return // directory add error return
		}
	}

	// apply event filter
	if w.filter != WatchOpAll && op&Op(w.filter) == 0 {
		return // filtered event return
	}

	var watchEvent = WatchEvent{
		At:       t,
		ID:       uuid.New(),
		BaseName: filepath.Base(name),
		AbsName:  name,
		Op:       op.String(),
		OpBits:   op,
	}

	w.eventFn(&watchEvent) // send event

	return
}

// whenever create if new directory,it is added
func (w *Watcher) addCreatedDirectoriesToWatcher(absName string) (err error) {

	// check if the created entry is a directory
	var fsFileInfo fs.FileInfo
	if fsFileInfo, err = os.Lstat(absName); err != nil {
		// ignore ENOENT because the entry may already have been removed
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		} else {
			err = perrors.Errorf("os.Lstat: %w", err)
		}
		return // os.Stat failed, do not know if its a directory
	}
	if !fsFileInfo.IsDir() {
		return // not a directory return
	}

	// add the newly created directory to the list of watched directories
	if err = w.WatcherShim.Add(absName); perrors.IsPF(&err, "watcher.Add: %w", err) {
		return
	}

	return
}

func (w *Watcher) shutdownOnErr(errp *error) {
	if *errp == nil || w.WatcherShim.ID == 0 {
		return
	}
	w.Shutdown()
}
