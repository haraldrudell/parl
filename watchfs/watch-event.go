/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/haraldrudell/parl"
)

// WatchEvent contains data about a file system event
// ID is unique for the event
type WatchEvent struct {
	// At is a time with ns resolution in the time.Local print format
	At time.Time
	// ID uniquely identifies the event, [16]byte
	ID uuid.UUID
	// BaseName filename.ext
	BaseName string
	// AbsName: /absolute-path/filename.ext
	AbsName string
	// Op is an or-strig of one or more file-system operations:
	// CREATE or
	// CREATE|REMOVE|WRITE|RENAME|CHMOD
	Op     string
	OpBits Op
}

func NewWatchEvent(fsnotifyEvent *fsnotify.Event, now time.Time, wa *Watcher) (watchEvent *WatchEvent) {
	return &WatchEvent{
		At:       now,
		ID:       uuid.New(),
		BaseName: filepath.Base(fsnotifyEvent.Name),
		AbsName:  fsnotifyEvent.Name,
		Op:       NewOp(fsnotifyEvent.Op).String(), // ensures the fswatch.Op exists
		OpBits:   NewOp(fsnotifyEvent.Op),
	}
}

func (we *WatchEvent) StringBase() (s string) {
	return we.string(false)
}

func (we *WatchEvent) String() (s string) {
	return we.string(true)
}

func (we *WatchEvent) string(printAbs bool) string {
	IDstring := we.ID.String()
	var name string
	if printAbs {
		name = we.AbsName
	} else {
		name = we.BaseName
	}
	return fmt.Sprintf("%s uuid: %s %s %s", // event: %#v",
		parl.Short(we.At),          // 220506_08:03:53-07
		IDstring[len(IDstring)-4:], // just the last 4 characters: uuid: b62f
		we.Op,                      // CREATE
		name,                       // a.txt or absoute path /…/a.txt
	)
}
