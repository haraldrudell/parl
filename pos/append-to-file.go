/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"io/fs"
	"os"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
)

// open flags: Linux man 2 open
const (
	// when created, output file permissions is user-read/write
	FilePermUrw os.FileMode = 0o600 // rw- --- ---
	// flags for os.OpenFile: write-only, append to if existing
	openFlagsCreateOrAppend =
	// if filename does not exist, create it as regular file
	os.O_CREATE |
		// append mode: any write is at end of file
		os.O_APPEND |
		// write-only: read methods return error
		os.O_WRONLY
	openFlagsNew = os.O_CREATE | os.O_APPEND | os.O_WRONLY |
		// used with O_CREATE, file must not exist.
		os.O_EXCL
	openFlagsLog = os.O_CREATE | os.O_APPEND | os.O_WRONLY |
		// each write commits data to the underlying hardware prior to return
		os.O_SYNC
	openFlagsExist             = os.O_APPEND | os.O_WRONLY
	FilePermUw     os.FileMode = 0o200 // -w- --- ---
	// for directories
	FilePermUrwx os.FileMode = 0o700 // rww --- ---
)

// AppendToFile is [os.Open] for a write-only file that is created or appended to
//   - use is an output file without data loss
//   - a created file has permissions urw
//   - if file is not on NFS, append is thread-safe
//   - other functions:
//   - — [os.Open]: opens an existing file for reading only
//   - — [os.Create]: opens or creates a file for reading and writing that is truncated if it exists
//   - — [pos.NewFile]: write-only file that must not exist
//   - — [pos.LogFile]: a created or appended to write-only file committing on every write
func AppendToFile(filename string, perms ...fs.FileMode) (osFile *os.File, err error) {
	var perm fs.FileMode
	if len(perms) > 0 {
		perm = perms[0]
	} else {
		perm = FilePermUrw
	}

	if osFile, err = os.OpenFile(filename, openFlagsCreateOrAppend, perm); err != nil {
		err = perrors.ErrorfPF("OpenFile: %w", err)
	}

	return
}

// AppenFile appends data to write-only file
//   - thread-safe if not on nfs
//   - default perms on create: urw
//   - similar to func os.WriteFile(name string, data []byte, perm os.FileMode) error
func AppendFile(filename string, data []byte, perms ...fs.FileMode) (err error) {
	var osFile *os.File
	if osFile, err = AppendToFile(filename, perms...); err != nil {
		return
	} else if _ /*n*/, err = osFile.Write(data); perrors.IsPF(&err, "File.Write %w", err) {
		return
	}
	parl.Close(osFile, &err)

	return
}

// AppenFile appends data to write-only file
//   - s: typically ending with newline
//   - thread-safe if not on nfs
//   - default perms on create: urw
//   - similar to func os.WriteFile(name string, data []byte, perm os.FileMode) error
func AppendFileString(filename, s string, perms ...fs.FileMode) (err error) {
	var osFile *os.File
	if osFile, err = AppendToFile(filename, perms...); err != nil {
		return
	} else if _ /*n*/, err = osFile.WriteString(s); perrors.IsPF(&err, "File.Write %w", err) {
		return
	}
	parl.Close(osFile, &err)

	return
}

// NewFile is [os.Open] for a write-only file that must exist
//   - if file is not on NFS, append is thread-safe
//   - other functions:
//   - — [os.Open]: opens an existing file for reading only
//   - — [os.Create]: opens or creates a file for reading and writing that is truncated if it exists
//   - — [pos.AppendToFile]: write-only file that is created or appended to
//   - — [pos.LogFile]: a created or appended to write-only file committing on every write
func AppendExisting(filename string) (osFile *os.File, err error) {
	if osFile, err = os.OpenFile(filename, openFlagsExist, FilePermUrw); err != nil {
		err = perrors.ErrorfPF("OpenFile: %w", err)
	}

	return
}

// NewFile is [os.Open] for a write-only file that must not exist
//   - use is an output file without data loss
//   - a created file has permissions urw
//   - if file is not on NFS, append is thread-safe
//   - other functions:
//   - — [os.Open]: opens an existing file for reading only
//   - — [os.Create]: opens or creates a file for reading and writing that is truncated if it exists
//   - — [pos.AppendToFile]: write-only file that is created or appended to
//   - — [pos.LogFile]: a created or appended to write-only file committing on every write
func NewFile(filename string) (osFile *os.File, err error) {
	if osFile, err = os.OpenFile(filename, openFlagsNew, FilePermUrw); err != nil {
		err = perrors.ErrorfPF("OpenFile: %w", err)
	}

	return
}

// LogFile is [os.Open] for a write-only file that is created or appended to
//   - use is an output file without data loss
//   - a created file has permissions urw
//   - if file is not on NFS, append is thread-safe
//   - written data is committed on every write
//   - other functions:
//   - — [os.Open]: opens an existing file for reading only
//   - — [os.Create]: opens or creates a file for reading and writing that is truncated if it exists
//   - — [pos.AppendToFile]: write-only file that is created or appended to
//   - — [pos.NewFile]: write-only file that must not exist
func LogFile(filename string) (osFile *os.File, err error) {
	if osFile, err = os.OpenFile(filename, openFlagsLog, FilePermUrw); err != nil {
		err = perrors.ErrorfPF("OpenFile: %w", err)
	}

	return
}

// CloseRm closes osFile and deletes it if errp contains error
//   - deferrable
func CloseRm(filename string, osFile *os.File, errp *error) {

	// close the file
	parl.Close(osFile, errp)

	// if there was no error and Close was successful: do not delete the file
	if *errp == nil {
		return
	}

	// delete outfile
	if err := RemoveFile(filename); err != nil {
		*errp = perrors.AppendError(*errp, err)
	}
}

// RemoveFile deletes a file after changing its permissions to ensure uw
func RemoveFile(filename string) (err error) {

	// ensure write permission for filename
	var fileInfo fs.FileInfo
	if fileInfo, err = pfs.Stat(filename); err != nil {
		return
	} else if fileInfo.Mode()&FilePermUw == 0 {
		if err = os.Chmod(filename, fileInfo.Mode()|FilePermUw); perrors.IsPF(&err, "Chmod %w", err) {
			return
		}
	}

	// remove file
	if err = os.Remove(filename); err != nil {
		err = perrors.ErrorfPF("os.Remove %w", err)
	}

	return
}
