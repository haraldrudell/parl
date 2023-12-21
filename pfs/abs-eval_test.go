/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/punix"
)

func TestAbsEval(t *testing.T) {
	//t.Fail()
	var dir = t.TempDir()
	var symlink = filepath.Join(dir, "symlink")
	if err := os.Symlink(dir, symlink); err != nil {
		panic(err)
	}
	var symlinkAbsEval = func() (symlinkAbsEval string) {
		var err error
		if symlinkAbsEval, err = filepath.Abs(dir); err != nil {
			panic(err)
		} else if symlinkAbsEval, err = filepath.EvalSymlinks(dir); err != nil {
			panic(err)
		}
		return
	}()
	var wd = func() (wd string) {
		var err error
		if wd, err = os.Getwd(); err != nil {
			panic(err)
		}
		return
	}()
	var symlinkRel = func() (symlinkRel string) {
		var err error
		if symlinkRel, err = filepath.Rel(wd, symlink); err != nil {
			panic(err)
		}
		return
	}()
	var empty = ""

	var absPath string
	var err error

	absPath, err = AbsEval(symlinkRel)

	// absPath "/private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestAbsEval2996198706/001"
	t.Logf("absPath %q", absPath)

	// absPath should not have error
	if err != nil {
		t.Errorf("AbsEval err %s", perrors.Short(err))
	}

	// absPath should retruned evaluated, absolute symlink
	if absPath != symlinkAbsEval {
		t.Errorf("AbsPath %q exp %q", absPath, symlinkAbsEval)
	}

	absPath, err = AbsEval(empty)

	// absPath empty should return wd
	if err != nil {
		t.Errorf("AbsEval2 err %s", perrors.Short(err))
	}
	if absPath != wd {
		t.Errorf("AbsPath2 %q exp %q", absPath, wd)
	}
}

func TestAbsEvalBadSymLink(t *testing.T) {
	//t.Fail()
	var badPath = "/%"
	var dir = t.TempDir()
	var symlink = filepath.Join(dir, "symlink")
	if err := os.Symlink(badPath, symlink); err != nil {
		panic(err)
	}

	var absPath string
	var err error

	// non-existing path
	absPath, err = AbsEval(symlink)

	// absPath ""
	t.Logf("absPath %q", absPath)
	// err:
	// ‘*errorglue.errorStack *fmt.wrapError *fs.PathError syscall.Errno’
	// ‘EvalSymlinks lstat /opt/sw/parl/pfs/%:
	// no such file or directory at pfs.AbsEval()-abs-eval.go:30’
	t.Logf("err: ‘%s’ ‘%s’", errorglue.DumpChain(err), perrors.Short(err))
	// is not exist: false
	t.Logf("is not exist: %t", os.IsNotExist(err))
	// errno: ENOENT 2 0x2
	t.Logf("errno: %s", punix.ErrnoString("", err))

	if err == nil {
		t.Error("AbsEval expected error missing")
	} else if !punix.IsENOENT(err) {
		t.Errorf("AbsEval bad err %s expected ENOENT", perrors.Short(err))
	}
}

func TestAbsEvalBadPath(t *testing.T) {
	var badPath = "/%"

	var absPath string
	var err error

	// non-existing path
	absPath, err = AbsEval(badPath)

	// absPath ""
	t.Logf("absPath %q", absPath)
	// err:
	// ‘*errorglue.errorStack *fmt.wrapError *fs.PathError syscall.Errno’
	// ‘EvalSymlinks lstat /opt/sw/parl/pfs/%:
	// no such file or directory at pfs.AbsEval()-abs-eval.go:30’
	t.Logf("err: ‘%s’ ‘%s’", errorglue.DumpChain(err), perrors.Short(err))
	// is not exist: false
	t.Logf("is not exist: %t", os.IsNotExist(err))
	// errno: ENOENT 2 0x2
	t.Logf("errno: %s", punix.ErrnoString("", err))

	if err == nil {
		t.Error("AbsEval expected error missing")
	} else if !punix.IsENOENT(err) {
		t.Errorf("AbsEval bad err %s expected ENOENT", perrors.Short(err))
	}
}
