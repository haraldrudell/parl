/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"strconv"

	"github.com/fsnotify/fsnotify"
)

const WatchOpAll fsnotify.Op = 0 // Create|Write|Remove|Rename|Chmod
const (
	Create fsnotify.Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

type Op fsnotify.Op

var allBits = Create | Write | Remove | Rename | Chmod

func (op Op) String() (s string) {
	s = fsnotify.Op(op).String()
	if unknownBits := op &^ Op(allBits); unknownBits != 0 {
		s += "|x" + strconv.FormatInt(int64(unknownBits), 16)
	}
	return
}
