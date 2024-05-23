/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package watchfs provides a file-system watcher for Linux and macOS.
package watchfs

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
)

func TestWatcherShim(t *testing.T) {
	//t.Fail()
	var dir = t.TempDir()
	var wd = func() (wd string) {
		var err error
		if wd, err = os.Getwd(); err != nil {
			panic(err)
		}
		return
	}()
	// dirRel is relative and may contain symlinks
	var dirRel = func() (dirRel string) {
		var err error
		if dirRel, err = filepath.Rel(wd, dir); err != nil {
			panic(err)
		}
		return
	}()
	var dirAbsEval = func() (dirAbsEval string) {
		var err error
		if dirAbsEval, err = pfs.AbsEval(dir); err != nil {
			panic(err)
		}
		return
	}()
	var dirUnclean = dir + "/."
	var pathsExp = []string{dirAbsEval}

	// dirRel: "../../../../var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherShim3150408602/001"
	t.Logf("dirRel: %q", dirRel)
	// dirAbsEval: "/private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherShim3150408602/001"
	t.Logf("dirAbsEval: %q", dirAbsEval)
	// dirUnclean: "/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestWatcherShim3825273188/001/."
	t.Logf("dirUnclean: %q", dirUnclean)

	var noFieldp *WatcherShim
	var shimT = newShimTester()
	var err error
	var paths []string

	// Add() List() Shutdown() Watch()
	var watcherShim *WatcherShim = NewWatcherShim(noFieldp, shimT.eventFunc, shimT)

	// Watch should not error
	err = watcherShim.Watch()
	if err != nil {
		t.Errorf("Watch err: %s", perrors.Short(err))
	}

	// Add should not error
	err = watcherShim.Add(dirUnclean)
	if err != nil {
		t.Errorf("Add err: %s", perrors.Short(err))
	}

	// List should return path
	paths = watcherShim.List()
	if !slices.Equal(paths, pathsExp) {
		t.Errorf("List %v exp %v", paths, pathsExp)
	}

	// Add should return error after Shutdown
	watcherShim.Shutdown()
	err = watcherShim.Add(dir)
	// err watchfs.Add shim idle, error or shutdown at watchfs.(*WatcherShim).Add()-watcher-shim.go:142
	t.Logf("err %s", perrors.Short(err))
	if err == nil {
		t.Error("Add missing err")
	}
}

// shimTester provides errFn and eventFunc
type shimTester struct{}

// newShimTester returns errFn and eventFunc implementations
func newShimTester() (w *shimTester) {
	return &shimTester{}
}

// errFn for watcherShim
func (w *shimTester) AddError(err error) {
	panic(err)
}

// eventFunc for watcherShim
func (w *shimTester) eventFunc(name string, op Op, t time.Time) (err error) {
	panic(err)
}
