/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"os"
	"slices"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// [DirectoryOrder]
	NoSorting = false
)

// directoryOrder returns unsorted entries, most with deferred [fs.FileInfo]
func ReadDir(abs string, sort ...bool) (entries []fs.DirEntry, err error) {
	// open the directory
	var osFile *os.File
	if osFile, err = os.Open(abs); perrors.Is(&err, "os.Open %w", err) {
		return
	}
	defer parl.Close(osFile, &err)

	// reads in directory order
	//	- returns an array of interface-value pointers, meaning
	//		each DirEntry is a separate allocation
	//	- for most modeTypes, Type is availble without lstat invocation
	//	- lstat is then deferred until Info method is invoked.
	//		Every time such deferring Info is invoked, lstat is executed
	if entries, err = osFile.ReadDir(allNamesAtOnce); perrors.Is(&err, "os.File.ReadDir %w", err) {
		return
	}

	// if NoSorting present, do not sort
	if len(sort) > 0 && !sort[0] {
		return
	}

	// sort by 8-bit characters
	slices.SortFunc(entries, compareDirEntry)

	return
}
