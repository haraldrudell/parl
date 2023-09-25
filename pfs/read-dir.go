/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"os"
	"sort"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	allNamesAtOnce = -1
)

// readDir reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDir(path string) (entryBasenames []string, err error) {

	var osFile *os.File
	if osFile, err = os.Open(path); perrors.Is(&err, "os.Open %w", err) {
		return
	}
	defer parl.Close(osFile, &err)
	// reads in directory order
	if entryBasenames, err = osFile.Readdirnames(allNamesAtOnce); perrors.Is(&err, "osFile.Readdirnames %w", err) {
		return
	}
	sort.Strings(entryBasenames)
	return
}
