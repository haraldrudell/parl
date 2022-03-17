/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License

go test -v github.com/haraldrudell/netter/parlfs
*/

package parlfs

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const (
	userPerm   os.FileMode = 0700
	uniqueName             = "TestWalkSym"
	rootName               = "root"
	root2Name              = "root2"
	dirName                = "dir"
	dir2Name               = "dir2"
	targetName             = "target"
	symName                = "sym"
	file1Name              = "file1"
	file2Name              = "file2"
)

func TestWalkSym(t *testing.T) {
	/*
		symlink test cases:
		1. repeated symlink makes no difference
		2. symlink to separate root
		3. symlink to parent
		4. symlink to root subdirectory

		./root/sym1 -> ./root2/dir2
		./root/sym2 -> ./root2/dir2
		./root/sym3 -> ./root/dir
		./root/dir/file1

		./root2/dir2/sym4 -> ./root2
		./root2/dir2/file2
	*/
	uniqueDir := filepath.Join(os.TempDir(), uniqueName)
	root := filepath.Join(uniqueDir, rootName)
	//panic(errors.New(uniqueDir))
	dir := filepath.Join(uniqueDir, rootName, dirName)
	dir2 := filepath.Join(uniqueDir, root2Name, dir2Name)
	for _, d := range []string{dir, dir2} {
		if err := os.MkdirAll(d, userPerm); err != nil {
			t.Logf("os.MkdirAll failed: %s\n", d)
			panic(err)
		}
	}
	for _, f := range []string{filepath.Join(dir, file1Name), filepath.Join(dir2, file2Name)} {
		file, err := os.Create(f)
		if err != nil {
			t.Logf("os.Create failed: %s\n", f)
			panic(err)
		}
		if err := file.Close(); err != nil {
			t.Logf("file.Close failed: %s\n", f)
			panic(err)
		}
	}
	oldnames := []string{
		filepath.Join(uniqueDir, root2Name, dir2Name),
		filepath.Join(uniqueDir, root2Name, dir2Name),
		filepath.Join(uniqueDir, rootName, dirName),
		filepath.Join(uniqueDir, root2Name),
	}
	for i := 0; i < 4; i++ {
		n := i + 1
		oldname := oldnames[i]
		newname := filepath.Join(uniqueDir, rootName, symName+strconv.Itoa(n))
		if n == 4 {
			newname = filepath.Join(uniqueDir, root2Name, dir2Name, symName+strconv.Itoa(n))
		}
		if err := os.Symlink(oldname, newname); err != nil && !os.IsExist(err) {
			t.Logf("os.Symlink failed: oldname: %s newname: %s\n", oldname, newname)
			panic(err)
		}
	}
	t.Logf("Directory: %s\n", uniqueDir)

	// WalkFunc
	var _ filepath.WalkFunc
	walkFn := func(path string, info os.FileInfo, err error) error {
		t.Log(path)
		return err
	}

	// filepath.Walk
	t.Log("filepath.Walk:")
	if err := filepath.Walk(root, walkFn); err != nil {
		t.Log("filepath.Walk failed")
		panic(err)
	}

	/*
		_ = symwalk.Walk
	*/
	/*
		// symwalk.Walk
		t.Log("symwalk.Walk:")
		if err := symwalk.Walk(root, walkFn); err != nil { // github.com/facebookgo/symwalk
			t.Log("symwalk.Walk failed")
			panic(err)
		}
	*/
	// parlfs.Walk
	t.Log("parlfs.Walk:")
	if err := Walk(root, walkFn); err != nil {
		t.Log("parlfs.Walk failed")
		panic(err)
	}
}
