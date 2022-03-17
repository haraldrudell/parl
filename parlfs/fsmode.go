/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlfs

import (
	"fmt"
	"io/fs"
	"strings"
)

type FileMode struct {
	fs.FileMode
}

const (
	// AllModeBits is all known github.com/fsnotify/fsevents.EventFlags
	AllModeBits fs.FileMode = fs.ModeDir | fs.ModeAppend | fs.ModeExclusive |
		fs.ModeTemporary | fs.ModeSymlink | fs.ModeDevice | fs.ModeNamedPipe |
		fs.ModeSocket | fs.ModeSetuid | fs.ModeSetgid | fs.ModeCharDevice |
		fs.ModeSticky | fs.ModeIrregular
	fsp = "fs."
)

type modeBit struct {
	fs.FileMode
	name string
	ch   string
	desc string
}

var modeBitList = []modeBit{
	{fs.ModeDir, "ModeDir", "d", "is a directory"},
	{fs.ModeAppend, "ModeAppend", "a", "append-only"},
	{fs.ModeExclusive, "ModeExclusive", "l", "exclusive use (lower case L)"},
	{fs.ModeTemporary, "ModeTemporary", "T", "temporary file; Plan 9 only"},
	{fs.ModeSymlink, "ModeSymlink", "L", "symbolic link"},
	{fs.ModeDevice, "ModeDevice", "D", "device file"},
	{fs.ModeNamedPipe, "ModeNamedPipe", "p", "named pipe (FIFO)"},
	{fs.ModeSocket, "ModeSocket", "S", "Unix domain socket"},
	{fs.ModeSetuid, "ModeSetuid", "u", "setuid"},
	{fs.ModeSetgid, "ModeSetgid", "g", "setgid"},
	{fs.ModeCharDevice, "ModeCharDevice", "c", "Unix character device, when ModeDevice is set"},
	{fs.ModeSticky, "ModeSticky", "t", "sticky"},
	{fs.ModeIrregular, "ModeIrregular", "?", "non-regular file; nothing else is known about this file"},
}

func GetModeBitValues() (s string) {
	var s1 []string
	for _, bit := range modeBitList {
		s1 = append(s1, fmt.Sprintf("%s %s %s: %s",
			FileMode{bit.FileMode}.Hex8(),
			bit.name,
			bit.ch, bit.desc))
	}
	return strings.Join(s1, "\n")
}

// Check returns hex string of unknown bits, empty string if no unknown bits
func (fm FileMode) Check() (str string) {
	unknownFlags := fm.FileMode & ^AllModeBits & ^fs.ModePerm
	if unknownFlags == 0 {
		return
	}
	return FileMode{unknownFlags}.Hex()
}

func (fm FileMode) Hex8() (s string) {
	d := uint32(fm.FileMode)
	return fmt.Sprintf("0x%08x", d)
}

func (fm FileMode) Hex() (s string) {
	d := uint32(fm.FileMode)
	return fmt.Sprintf("0x%x", d)
}

func (fm FileMode) Or() (s string) {
	fsFileMode := fm.FileMode
	if fsFileMode == 0 {
		return "0"
	}
	var s1 []string
	for _, bit := range modeBitList {
		if fsFileMode&bit.FileMode != 0 {
			s1 = append(s1, fsp+bit.name)
		}
	}
	return strings.Join(s1, "|")
}

func (fm FileMode) Ch() (s string) {
	fsFileMode := fm.FileMode
	for _, bit := range modeBitList {
		if fsFileMode&bit.FileMode != 0 {
			s += bit.ch
		}
	}
	return
}

func (fm FileMode) String() (s string) {
	fsFileMode := fm.FileMode
	if fsFileMode == 0 {
		return "0"
	}
	s = fm.Hex() + fm.Or()
	s2 := fm.Check()
	if s2 != "" {
		s += "-bad:" + s2
	}
	return
}
