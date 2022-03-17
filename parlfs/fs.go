/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlfs

import (
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/haraldrudell/parl"
)

const (
	mvCommand                                    = "mv"
	mvNoClobber                                  = "-n"
	defaultDirMode                               = 0o700
	IsDirectoryNonExistentIsError IsDirectoryArg = 1 << iota
	IsDirectoryNotDirIsError
)

type IsDirectoryArg byte

func Mv(src, dest string, outConsumer func(string)) (err error) {
	parl.Debug("Mv src: %s dest: %s\n", src, dest)
	var bytes []byte
	bytes, err = exec.Command(mvCommand, mvNoClobber, src, dest).CombinedOutput()
	if len(bytes) != 0 && outConsumer != nil {
		outConsumer(string(bytes))
	}
	if err != nil {
		err = parl.Errorf("exec.Command mv: '%w'", err)
	}
	return
}

func Stat(path string) (fileInfo fs.FileInfo, err error) {
	if fileInfo, err = os.Stat(path); err != nil {
		err = parl.Errorf("os.Stat: '%w'", err)
	}
	return
}

// fileInfo: does exist, nil: does not exist
func Exists(path string) (fileInfo fs.FileInfo /* interface */) {
	var err error
	fileInfo, err = os.Stat(path)
	if err == nil {
		return // does exist: fileInfo
	}
	if os.IsNotExist(err) {
		return // does not exist: nil
	}
	panic(parl.Errorf("os.Stat: '%w'", err))
}

func IsDirectory(path string, flags IsDirectoryArg) (isDirectory bool, err error) {
	fileInfo := Exists(path)
	if fileInfo == nil {
		if flags&IsDirectoryNonExistentIsError != 0 {
			err = parl.Errorf("Does not exist: %s", path)
		}
		return
	}
	if isDirectory = fileInfo.IsDir(); !isDirectory && IsDirectoryNotDirIsError != 0 {
		err = parl.Errorf("Not directory: %s", path)
	}
	return
}

func IsEmptyDirectory(path string) (isEmpty bool) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(parl.Errorf("os.Open: '%w'", err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(parl.Errorf("file.Close: '%w'", err))
		}
	}()
	_, err = file.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			return true
		}
		panic(parl.Errorf("Readdirnames: '%w'", err))
	}
	return false // directory has files
}

// use 0 for default file mode owner rwx
func EnsureDirectory(directory string, dirMode fs.FileMode) {
	fileInfo, err := os.Stat(directory)
	if err == nil {
		if !fileInfo.IsDir() {
			panic(parl.Errorf("Is not directory: %s", directory))
		}
		return // does exist, is directory
	}
	if !os.IsNotExist(err) {
		panic(parl.Errorf("os.Stat: '%w'", err))
	}
	if dirMode == 0 {
		dirMode = defaultDirMode
	}
	if err = os.MkdirAll(directory, dirMode); err != nil {
		panic(parl.Errorf("os.MkdirAll: '%w'", err))
	}
}

var i = 0

func MoveOrMerge(src, dest string, outConsumer func(string)) (err error) {
	i++
	ii := i
	parl.Debug("MoveOrMerge %d src: %s dest: %s\n", ii, src, dest)

	// if dest does not exist, create dest’s parents and move src
	fileInfo := Exists(dest)
	parl.Debug("MoveOrMerge %d Exists complete\n", ii)
	if fileInfo == nil { // destination does not exist: move the source
		EnsureDirectory(filepath.Dir(dest), 0) // ensure parent directory exists
		parl.Debug("MoveOrMerge %d outcome: mv\n", ii)
		return Mv(src, dest, outConsumer) // move the directory
	}

	// ensure src and dest are both directories
	var srcInfo fs.FileInfo
	if srcInfo, err = os.Stat(src); err != nil {
		parl.Debug("MoveOrMerge %s os.Stat\n", ii)
		return
	}
	if !srcInfo.IsDir() {
		parl.Debug("MoveOrMerge %d src !IsDir\n", ii)
		return parl.Errorf("MoveOrMerge: source is file, dest exists: %s", src)
	}
	if !fileInfo.IsDir() {
		parl.Debug("MoveOrMerge %d dst !IsDir\n", ii)
		return parl.Errorf("MoveOrMerge: source is directory, dest is not: %s", src)
	}

	// merge src/* into dest
	parl.Debug("MoveOrMerge %d merge\n", ii)
	reader := NewDirStream(src, 0)
	defer reader.Shutdown()
	for entryPack := range reader.Results {
		baseName := entryPack.Name()
		if err = MoveOrMerge(
			filepath.Join(src, baseName),
			filepath.Join(dest, baseName),
			outConsumer); err != nil {
			parl.Debug("MoveOrMerge %d entry err\n", ii)
			return
		}
	}
	if IsEmptyDirectory(src) {
		parl.Debug("EMPTY: %d %s\n", ii, src)
		err = os.Remove(src)
	} else {
		parl.Debug("NOTEMPTY: %d %s\n", ii, src)
	}
	return
}
