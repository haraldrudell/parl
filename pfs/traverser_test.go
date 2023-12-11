/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

// tests traversing single entry
func TestTraverser(t *testing.T) {
	var tempDir = t.TempDir()
	// path is unclean providedPath
	var path = filepath.Join(tempDir, ".")
	var expAbs = func() (abs string) {
		var e error
		if abs, e = filepath.Abs(tempDir); e != nil {
			panic(e)
		} else if abs, e = filepath.EvalSymlinks(abs); e != nil {
			panic(e)
		}
		return
	}()
	var expNo1 = uint64(1)
	var expBase = filepath.Base(tempDir)

	var result ResultEntry

	var traverser = NewTraverser(path)

	// first Next should be dir
	result = traverser.Next()
	if result.ProvidedPath != path {
		t.Errorf("Next path %q exp %q", result.ProvidedPath, path)
	}
	if result.Abs != expAbs {
		t.Errorf("Next Abs %q exp %q", result.Abs, expAbs)
	}
	if result.No != expNo1 {
		t.Errorf("Next No %d exp %d", result.No, expNo1)
	}
	if result.Err != nil {
		t.Errorf("Next Err: %s", perrors.Short(result.Err))
	}
	if result.SkipEntry == nil {
		t.Error("Next SkipEntry nil")
	}
	if n := result.Name(); n != expBase {
		t.Errorf("Next Name %q exp %q", n, expBase)
	}

	// NextNext should be end
	result = traverser.Next()
	if !result.IsEnd() {
		t.Error("NextNext not end")
	}
}

// tests traversing directory
func TestTraverserDir(t *testing.T) {
	var tempDir = t.TempDir()
	var fileBase = "file"
	var filename = filepath.Join(tempDir, fileBase)
	if e := os.WriteFile(filename, nil, 0700); e != nil {
		panic(e)
	}

	var result ResultEntry

	var traverser = NewTraverser(tempDir)

	// Next should be tempDir
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next err: %s", result.Err)
	}
	if result.ProvidedPath != tempDir {
		t.Errorf("Next ProvidedPath %q exp %q", result.ProvidedPath, tempDir)
	}

	// NextNext should be filename
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNext err: %s", result.Err)
	}
	if result.ProvidedPath != filename {
		t.Errorf("NextNext ProvidedPath %q exp %q", result.ProvidedPath, filename)
	}

	// NextNextNext should be end
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNextNext err: %s", result.Err)
	}
	if !result.IsEnd() {
		t.Error("NextNextNext not end")
	}
}

// tests returning an error
func TestTraverserError(t *testing.T) {
	var path = "%"

	var result ResultEntry

	var traverser = NewTraverser(path)

	// first Next should be err
	result = traverser.Next()
	if result.Err == nil {
		t.Error("Next Err nil")
	}
}

// tests skip of entry
func TestTraverserSkip(t *testing.T) {
	var tempDir = t.TempDir()
	var fileBase = "file"
	var filename = filepath.Join(tempDir, fileBase)
	if e := os.WriteFile(filename, nil, 0700); e != nil {
		panic(e)
	}

	var result ResultEntry

	var traverser = NewTraverser(tempDir)

	// Next should be tempDir
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next err: %s", result.Err)
	}
	if result.ProvidedPath != tempDir {
		t.Errorf("Next ProvidedPath %q exp %q", result.ProvidedPath, tempDir)
	}
	if result.SkipEntry == nil {
		t.Fatalf("Next result.SkipEntry nil")
	}
	result.Skip()

	// NextNext should be end
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNext err: %s", result.Err)
	}
	if !result.IsEnd() {
		t.Error("NextNext not end")
	}
}

// tests that in-root symlinks are ignored
func TestTraverserInRootSymlink(t *testing.T) {
	var tempDir = t.TempDir()
	var symlink = filepath.Join(tempDir, "symlink")
	if e := os.Symlink(tempDir, symlink); e != nil {
		panic(e)
	}
	var cleanTempDir = func() (abs string) {
		var err error
		if abs, err = filepath.Abs(tempDir); err != nil {
			panic(err)
		} else if abs, err = filepath.EvalSymlinks(abs); err != nil {
			panic(err)
		}
		return
	}()

	var result ResultEntry

	var traverser = NewTraverser(tempDir)

	// Next should be tempDir
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next err: %s", result.Err)
	}
	if result.ProvidedPath != tempDir {
		t.Errorf("Next ProvidedPath %q exp %q", result.ProvidedPath, tempDir)
	}
	if result.SkipEntry == nil {
		t.Fatalf("Next result.SkipEntry nil")
	}

	// NextNext should be symlink
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNext err: %s", result.Err)
	}
	if result.ProvidedPath != symlink {
		t.Errorf("NextNext ProvidedPath %q exp %q", result.ProvidedPath, symlink)
	}
	if result.Abs != cleanTempDir {
		t.Errorf("NextNext Abs %q exp %q", result.Abs, cleanTempDir)
	}

	// NextNextNext should be end
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNextNext err: %s", result.Err)
	}
	if !result.IsEnd() {
		t.Errorf("NextNextNext not end path: %q", result.ProvidedPath)
	}
}

// tests that disparate symlinks are followed
func TestTraverserDisparateSymlink(t *testing.T) {
	var tempDir = t.TempDir()
	var tempDir2 = t.TempDir()
	var cleanDir2 = func() (abs string) {
		var err error
		if abs, err = filepath.Abs(tempDir2); err != nil {
			panic(err)
		} else if abs, err = filepath.EvalSymlinks(abs); err != nil {
			panic(err)
		}
		return
	}()
	var symlink = filepath.Join(tempDir, "symlink")
	if e := os.Symlink(tempDir2, symlink); e != nil {
		panic(e)
	}

	var result ResultEntry

	var traverser = NewTraverser(tempDir)

	// Next should be tempDir
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next err: %s", result.Err)
	}
	if result.ProvidedPath != tempDir {
		t.Errorf("Next ProvidedPath %q exp %q", result.ProvidedPath, tempDir)
	}

	// NextNext should be symlink
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNext err: %s", result.Err)
	}
	if result.ProvidedPath != symlink {
		t.Errorf("NextNext ProvidedPath %q exp %q", result.ProvidedPath, symlink)
	}
	if result.Abs != cleanDir2 {
		t.Errorf("NextNext Abs %q exp %q", result.Abs, cleanDir2)
	}

	// Next3 should be cleanDir2
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next3 err: %s", result.Err)
	}
	if result.IsEnd() {
		t.Error("Next3 IsEnd")
	}
	if result.ProvidedPath != cleanDir2 {
		t.Errorf("Next3 ProvidedPath %q exp %q", result.ProvidedPath, cleanDir2)
	}

	// Next4 should be end
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next4 err: %s", result.Err)
	}
	if !result.IsEnd() {
		t.Error("Next4 not end")
	}
}

func TestTraverserBrokenSymlink(t *testing.T) {
	//t.Fail()
	var tempDir = t.TempDir()
	var brokenLink = filepath.Join(tempDir, "brokenLink")
	if e := os.Symlink(filepath.Join(tempDir, "nonExistent"), brokenLink); e != nil {
		panic(e)
	}

	var result ResultEntry

	var traverser = NewTraverser(brokenLink)

	// Next should be err
	result = traverser.Next()

	// result.Err: pfs.Load filepath.EvalSymlinks
	// lstat /private/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/TestTraverserBrokenSymlink814588820/001/nonExistent:
	// no such file or directory at pfs.(*Root2).Load()-root2.go:51
	t.Logf("result.Err: %s", perrors.Short(result.Err))

	if result.Err == nil {
		t.Error("Next missing err")
	} else if !errors.Is(result.Err, fs.ErrNotExist) {
		t.Errorf("Next err not fs.ErrNotExist: %s", perrors.Short(result.Err))
	}
}

func TestTraverserObsoletingSymlink(t *testing.T) {
	var tempDir = t.TempDir()
	var dir = filepath.Join(tempDir, "dir")
	if e := os.Mkdir(dir, 0700); e != nil {
		panic(e)
	}
	var obsoletingLink = filepath.Join(dir, "obsoletingLink")
	if e := os.Symlink(tempDir, obsoletingLink); e != nil {
		panic(e)
	}
	var cleanTempDir = func() (abs string) {
		var err error
		if abs, err = filepath.Abs(tempDir); err != nil {
			panic(err)
		} else if abs, err = filepath.EvalSymlinks(abs); err != nil {
			panic(err)
		}
		return
	}()

	var result ResultEntry

	var traverser = NewTraverser(dir)

	// Next should be dir
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next err: %s", result.Err)
	}
	if result.ProvidedPath != dir {
		t.Errorf("Next ProvidedPath %q exp %q", result.ProvidedPath, dir)
	}

	// Next2 should be obsoletingLink
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("Next2 err: %s", result.Err)
	}
	if result.IsEnd() {
		t.Error("Next2 end")
	}
	if result.ProvidedPath != obsoletingLink {
		t.Errorf("Next2 ProvidedPath %q exp %q", result.ProvidedPath, obsoletingLink)
	}
	if result.Abs != cleanTempDir {
		t.Errorf("Next2 Abs %q exp %q", result.Abs, cleanTempDir)
	}

	// NextNext should be cleanTempDir
	//	- dir/obsoletingLink links to parent tempDir
	//	- entries from dir ends
	//	- the new root is traversed
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNext err: %s", result.Err)
	}
	if result.IsEnd() {
		t.Error("NextNext end")
	}
	if result.ProvidedPath != cleanTempDir {
		t.Errorf("NextNext ProvidedPath %q exp %q", result.ProvidedPath, cleanTempDir)
	}

	// NextNextNext should end
	result = traverser.Next()
	if result.Err != nil {
		t.Errorf("NextNextNext err: %s", result.Err)
	}
	if !result.IsEnd() {
		t.Error("NextNextNext not end")
	}
}
