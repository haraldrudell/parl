/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/ptime"
	"github.com/haraldrudell/parl/punix"
)

/*
NewWatcher produces file-system events from a file-system entry and its child directories.
WatchEvents must be listened for on the .Events() channel until it closes.
Events that can be listened for are WatchOpAll (default: 0) Create Write Remove Rename Chmod.
errors must be listened for on the .Errors() channel until it closes.
Close the watcher by canceling the context or invoking .Shutdown().
// errFn must be thread-safe.
// eventFn must be thread-safe.

220505 github.com/fsnotify/fsnotify v1.5.4
220315 github.com/fsnotify/fsnotify v1.4.9 does not support macOS
220315 use the old github.com/fsnotify/fsevents v0.1.1
*/
func NewWatcher(
	directory string, filter Op,
	ignores *regexp.Regexp,
	eventFn func(event *WatchEvent),
	errFn func(err error)) (watcher *Watcher) {
	now := time.Now()
	var abs string
	var err error
	if abs, err = filepath.Abs(directory); err != nil {
		panic(perrors.Errorf("filepath.Abs: %q '%w'", directory, err))
	}

	// initialize watcher
	filter.fsnotifyOp() // check that fi;lter is valid
	watcher = NewWatcher0(ignores, eventFn, errFn)
	defer func() {
		if err != nil {
			watcher.Shutdown()
			panic(err)
		}
	}()
	watcher.abs = abs
	watcher.cleanDir = filepath.Clean(directory)
	watcher.filter = filter

	// insert our directory filter
	watcher.eventFn0 = eventFn
	watcher.eventFn = watcher.eventFilter

	// scan directory for all subdirectories
	// neither Linux and macOS can watch a directory tree so
	// each watched directory needs to be referenced
	if _, err = pfs.Dirs(directory, func(dir string) (err error) {
		if ignores != nil && ignores.MatchString(dir) {
			return
		}
		if err = watcher.watcher.Add(dir); err != nil {
			err = perrors.Errorf("watcher.Add: %w", err)
			return
		}
		return
	}); err != nil {
		return
	}
	d := time.Now().Sub(now).Round(100 * time.Millisecond)
	parl.Debug("%s WatchFS directories added in %s\n", ptime.NsLocal(now), d)
	return
}

func (w *Watcher) eventFilter(event *WatchEvent) {

	// if it is CREATE of a directory, add that directory to be watched, too
	if event.OpBits&Create != 0 {
		w.addCreatedDirectoriesToWatcher(event.AbsName)
	}

	// apply event filter
	if w.filter != WatchOpAll && event.OpBits&Op(w.filter) == 0 {
		return // filtered event exit
	}

	w.eventFn0(event) // send event
	return
}

func (w *Watcher) addCreatedDirectoriesToWatcher(absName string) {

	// check if the created entry is a directory
	var fsFileInfo fs.FileInfo
	var err error
	if fsFileInfo, err = os.Stat(absName); err != nil {
		// ignore ENOENT because the entry may already have been removed
		if !punix.IsENOENT(err) {
			w.errFn(perrors.Errorf("os.Stat: %w", err))
		}
		return // os.Stat failed, do not know if its a directory
	} else if !fsFileInfo.IsDir() {
		return // not a directory return
	}

	// add the newly created directory to the list of wathed directories
	if err = w.watcher.Add(absName); err != nil {
		w.errFn(perrors.Errorf("watcher.Add: %w", err))
		return // watcher.Add failed
	}

	return
}
