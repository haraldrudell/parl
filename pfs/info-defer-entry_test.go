/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"testing"
)

func TestInfoEntryWithFileInfo(t *testing.T) {
	// t.Fail()
	// var expType fs.FileMode = 0
	// var tempDir = t.TempDir()
	// var regularBase = "regular.txt"
	// var regularName = filepath.Join(tempDir, regularBase)
	// var content = []byte("abc")
	// var mode fs.FileMode = 0700
	// if e := os.WriteFile(regularName, content, mode); perrors.Is(&e, "os.WriteFile %w", e) {
	// 	panic(e)
	// }

	// // filepath.Abs always returns non-empty string
	// // if s, e := filepath.Abs(""); e != nil {
	// // 	t.Errorf("filepath.Abs('') err: %s", perrors.Short(e))
	// // } else {
	// // 	// filepath.Abs(''): "/opt/sw/parl/pfs"
	// // 	t.Logf("filepath.Abs(''): %q", s)
	// // }

	// var err error
	// var entry *InfoEntry

	// entry, err = AddDirEntry(regularName)
	// if err != nil {
	// 	t.Errorf("InfoEntryWithFileInfo err: %s", perrors.Short(err))
	// }
	// if typ := entry.Type(); typ != expType {
	// 	t.Errorf("Type %v exp %v", typ, expType)
	// }
	// if n := entry.Name(); n != regularBase {
	// 	t.Errorf("Name %q exp %q", n, regularBase)
	// }
}

func TestInfoEntryWithFileInfoSymlink(t *testing.T) {
	// var expType fs.FileMode = 0
	// var tempDir = t.TempDir()
	// var regularBase = "regular.txt"
	// var regularName = filepath.Join(tempDir, regularBase)
	// var content = []byte("abc")
	// var mode fs.FileMode = 0700
	// if e := os.WriteFile(regularName, content, mode); perrors.Is(&e, "os.WriteFile %w", e) {
	// 	panic(e)
	// }
	// var symlinkBase = "symlink.txt"
	// var symlinkName = filepath.Join(tempDir, symlinkBase)
	// if e := os.Symlink(regularName, symlinkName); perrors.Is(&e, "os.Symlink %w", e) {
	// 	panic(e)
	// }

	// var err error
	// var entry *InfoEntry

	// entry, err = AddDirEntry(symlinkName)
	// if err != nil {
	// 	t.Errorf("InfoEntryWithFileInfo err: %s", perrors.Short(err))
	// }
	// if typ := entry.Type(); typ != expType {
	// 	t.Errorf("Type %s exp %s", typ, expType)
	// }
	// if n := entry.Name(); n != regularBase {
	// 	t.Errorf("Name %q exp %q", n, regularBase)
	// }
}
