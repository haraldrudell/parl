/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"strconv"

	"github.com/fsnotify/fsnotify"
	"github.com/haraldrudell/parl/perrors"
)

// enum bit values for watchfs.Op are the same as fsnotify.Op
const (
	// Create|Write|Remove|Rename|Chmod
	WatchOpAll Op = 0
	Create        = Op(fsnotify.Create)
	Write         = Op(fsnotify.Write)
	Remove        = Op(fsnotify.Remove)
	Rename        = Op(fsnotify.Rename)
	Chmod         = Op(fsnotify.Chmod)
)

// Op allows callers to not import fsnotify dependency
// Op value is a bitfield of or of one or more Op enum bits
type Op uint32

// list of all watchfs.Op bit values
var opList = []Op{
	Create,
	Write,
	Remove,
	Rename,
	Chmod,
}

// allBits are all known bits used by fsnotify.Op values
var allBits = func() (bits uint32) {
	for _, op := range opList {
		bits |= uint32(op)
	}
	return
}()

// NewOp creates a watchfs.Op from an fsnotify.Op
func NewOp(o fsnotify.Op) (op Op) {

	// check for unknown bits
	if unknownBits := uint32(o) &^ allBits; unknownBits != 0 {
		panic(perrors.Errorf("NewOp unknown fsnotify.Op bits: 0x%s", strconv.FormatInt(int64(unknownBits), 16)))
	}

	return Op(o)
}

// fsnotifyOp converts a watchfs.Op to an fsnotify.Op
func (op Op) fsnotifyOp() (o fsnotify.Op) {

	// check for unknown bits
	if unknownBits := uint32(op) &^ allBits; unknownBits != 0 {
		panic(perrors.Errorf("fsnotifyOp unknown fsnotify.Op bits: 0x%s", strconv.FormatInt(int64(unknownBits), 16)))
	}

	return fsnotify.Op(op)
}

// Uint32 returns the uint 32 vaklue of the op bit field
func (op Op) Uint32() (value uint32) {
	return uint32(op)
}

// OpList returns the list of set bits Ops in op.
func (op Op) OpList() (ops []Op) {
	for _, o := range opList {
		if op&o != 0 {
			ops = append(ops, o)
		}
	}
	return
}

// HasOp determines if op has the o Op in its bit-encoded value.
func (op Op) HasOp(o Op) (hasOp bool) {
	return op&o != 0
}

func (op Op) String() (s string) {
	s = fsnotify.Op(op).String()
	if unknownBits := uint32(op) &^ allBits; unknownBits != 0 {
		s += "|x" + strconv.FormatInt(int64(unknownBits), 16)
	}
	return
}
