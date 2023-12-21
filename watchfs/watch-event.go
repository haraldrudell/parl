/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/haraldrudell/parl"
)

// WatchEvent contains data about a file system event
//   - ID is unique for the event
//   - WatchEvent is a value-container that as a local variable, function argument or result
//     is a tuple not causing allocation by using temporary stack storage
//   - taking the address of a &ResultEntry causes allocation
type WatchEvent struct {
	// At is a time with ns resolution in time.Local time-zone
	At time.Time
	// ID uniquely identifies the event, [16]byte
	ID uuid.UUID
	// BaseName “filename.ext”
	BaseName string
	// AbsName: “/absolute-path/filename.ext”
	AbsName string
	// Op is an or-string of one or more file-system operations:
	//	- CREATE or
	//	- CREATE|REMOVE|WRITE|RENAME|CHMOD
	Op string
	// OpBits is operation bitfield
	OpBits Op
}

func (e WatchEvent) Dump() (s string) {
	return parl.Sprintf("at: %s id: %s base: %q abs: %q op: %s %s",
		e.At.Format(parl.Rfc3339ns),
		e.ID,
		e.BaseName,
		e.AbsName,
		e.Op,
		e.OpBits,
	)

}

func (e WatchEvent) String() (s string) {
	var IDstring = e.ID.String()
	return fmt.Sprintf("%s uuid: %s %s %s", // event: %#v",
		parl.Short(e.At),           // 220506_08:03:53-07
		IDstring[len(IDstring)-4:], // just the last 4 characters: uuid: b62f
		e.Op,                       // CREATE
		e.BaseName,                 // a.txt or absoute path /…/a.txt
	)
}
