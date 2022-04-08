/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"path/filepath"
	"syscall"

	"github.com/haraldrudell/parl"
)

type DLEntry struct {
	RelDir      string // directory name that may begin with '.'
	AbsDir      string // absolute directory name
	FqPath      string // fully qualified path to entry
	fs.DirEntry        // .Name() .IsDir() .Type() .Info()
	Info        fs.FileInfo
	Stat        *syscall.Stat_t
}

func GetEntry(rel, abs string, entry fs.DirEntry, info fs.FileInfo, stat *syscall.Stat_t) (e *DLEntry) {
	return &DLEntry{RelDir: rel, AbsDir: abs, FqPath: filepath.Join(abs, entry.Name()), DirEntry: entry, Info: info, Stat: stat}
}

type EntryResult struct {
	*DLEntry
	Err error
}

func GetErrorResult(err error) (result *EntryResult) {
	if err == nil {
		panic(parl.Errorf("GetErrorResult with error nil"))
	}
	return &EntryResult{Err: err}
}
