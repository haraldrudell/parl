/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"os"
	"testing"

	"github.com/haraldrudell/parl/iters"
)

func TestNewDirEntryIterator(t *testing.T) {
	//t.Errorf("Logging on")

	var (
		err error
		wd  string
	)

	// Init() Cond() Next() Cancel()
	var iterator iters.Iterator[DirEntry]
	var _ = err

	// empty string path should error
	iterator = NewDirEntryIterator("")
	for dirEntry, _ := iterator.Init(); iterator.Cond(&dirEntry); {
		t.Logf("entry0: %q", dirEntry.Name())
		break
	}
	err = iterator.Cancel()
	if err == nil {
		t.Errorf("iterator.Cancel missing err")
	}

	// iteration values
	wd, err = os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd err %s", err)
	}
	// wd: "/opt/sw/parl/pfs"
	t.Logf("wd: %q", wd)
	iterator = NewDirEntryIterator(".")
	var breaker = false
	for dirEntry, _ := iterator.Init(); iterator.Cond(&dirEntry) && !breaker; {
		t.Logf("dirEntry.Name: %q", dirEntry.Name())
		t.Logf("dirEntry.ProvidedPath: %q", dirEntry.ProvidedPath)
		var abs string
		if abs, err = AbsEval(dirEntry.ProvidedPath); err != nil {
			t.Errorf("pfs.AbsEval err %s", err)
		}
		t.Logf("abs: %q", abs)
		t.Logf("dirEntry.IsDir: %t", dirEntry.IsDir())
		t.Logf("dirEntry.Type().IsRegular(): %t", dirEntry.Type().IsRegular())
		var fileInfo fs.FileInfo
		if fileInfo, err = dirEntry.Info(); err != nil {
			t.Errorf("dirEntry.Info err %s", err)
		}
		t.Logf("fileInfo.Size %d", fileInfo.Size())

		breaker = true
	}
	err = iterator.Cancel()
	if err != nil {
		t.Errorf("iterator.Cancel err %s", err)
	}
}
