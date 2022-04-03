/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WatchEvent struct {
	At        time.Time // ns resolution
	ID        uuid.UUID // unique event identifier [16]byte
	BaseName  string    // filename.ext
	AbsName   string    // /absolute-path/filename.ext
	CleanName string    // may-have-symlinks/filename.ext
	Op        string    // CREATE|REMOVE|WRITE|RENAME|CHMOD
	//fsnotify.Event           // Name: string path relative to directory Op: fsnotify.Op uint32
}

func (we WatchEvent) String() string {
	return fmt.Sprintf("%s uuid: %s base: %s", // event: %#v",
		we.At,
		we.ID,
		we.BaseName,
		//we.Event,
	)
}
