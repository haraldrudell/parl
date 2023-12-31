/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"unicode"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/punix"
)

const (
	// path segment “.local”.
	// “~/.local/share” is a standardized directory on Linux
	dotLocalDir = ".local"
	// path segement “share”.
	// “~/.local/share” is a standardized directory on Linux
	shareDir = "share"
	// mode for created directories
	urwx os.FileMode = 0700
)

// for testing
var homeDirHook string

// AppDirectory manages a per-user writable app-specific directory
type AppDirectory struct {
	// app name like “myapp”
	App string
	// absolute clean symlink-free path if app-directory exists
	//	- macOS: “/Users/user/.local/share/myapp”
	//	- Linux: “/home/user/.local/share/myapp”
	abs atomic.Pointer[string]
}

// NewAppDir returns a writable directory object in the user’s home directory
//   - appName: application name like “myapp”
//     Unicode letters and digits
//   - directory is “~/.local/share/[appName]”
//   - parent directory is based on the running process’ owner
//   - does not rely on environment variables
//
// Usage:
//
//	var appDir = NewAppDir("myapp")
//	if err = appDir.EnsureDir(); err != nil {…
//	var knownToExistAbsCleanNoSymlinksNeverErrors = appDir.Directory()
func NewAppDir(appName string) (appd *AppDirectory) { return &AppDirectory{App: appName} }

// best-effort single-value absolute clean possibly symlink-free directory
//   - returns an absolute path whether the directory exists or not
//   - if directory exists, absolute clean symlink-free, otherwise absolute clean
//   - Directory may panic from errors that are returned by [AppDirectory.EnsureDir] or
//     [AppDirectory.Path].
//     To avoid panics, invoke those methods first.
//
// Usage:
//
//	var dir = NewAppDir("myapp").Directory()
func (d *AppDirectory) Directory() (abs string) {
	var isNotExist bool
	var err error
	if abs, isNotExist, err = d.Path(); err != nil && !isNotExist {
		panic(err) // some error
	}
	return
}

// EnsureDir ensures the directory exists
func (d *AppDirectory) EnsureDir() (err error) {

	// get path while checking if already exists
	var abs string
	var isNotExist bool
	if abs, isNotExist, err = d.Path(); err == nil {
		return // directory already exists return
	} else if !isNotExist {
		return // some error
	}

	// MkDirAll begins with stat to see if path exists
	if err = os.MkdirAll(abs, urwx); perrors.IsPF(&err, "os.MkdirAll: %w", err) {
		return
	}
	// update d.abs
	_, _, err = d.eval(abs)

	return
}

// Path returns best-effort absolute clean path
//   - if the app-directory exists, abs is also symlink-free
//   - outcomes:
//   - — err: nil: abs is absolute clean symlink-free, app directory exists
//   - — isNotExist: true, err: non-nil: app directory does not eixst.
//     abs is absolute clean.
//     err is errno ENOENT
//   - — err: non-nil, isNotExist: false: some error
//   - —
//   - macOS: “/Users/user/.local/share/myapp”
//   - Linux: “/home/user/.local/share/myapp”
//   - note: symlinks can only be evaled if a path exists
func (d *AppDirectory) Path() (abs string, isNotExist bool, err error) {

	// if already present
	if ap := d.abs.Load(); ap != nil {
		abs = *ap
		return // success: already has abs, directory exists return
	}

	// check appName
	var appName string
	if appName, err = d.checkAppName(); err != nil {
		return // bad appName return
	}

	// get user’s home directory
	var homeDir string
	if h := homeDirHook; h == "" {
		if homeDir, err = UserHome(); err != nil {
			return // failure to obtain home directory return
		}
	} else {
		homeDir = h
	}

	// get app directory’s parent
	//	- absolute, maybe unclean, maybe symlinks
	var parentDir = filepath.Join(homeDir, dotLocalDir, shareDir)

	// get app directory
	//	- absolute, maybe unclean, maybe symlinks
	var a = filepath.Join(parentDir, appName)

	// try to unsymlink app directory
	if abs, isNotExist, err = d.eval(a); err == nil {
		return // app directory exists success return
	} else if !isNotExist {
		return // some error return
	}
	// err is non-nil, isNotExist true

	// try to unsymlink parent directory
	if p, e := pfs.AbsEval(parentDir); e != nil {
		if punix.IsENOENT(e) {
			abs = a
			return // parent no exist either, return isNotExist result
		}
		err = e // some new error
		isNotExist = false
		return // return error from parent directory
	} else {
		// use th evealed parent directory
		abs = filepath.Clean(filepath.Join(p, appName))
	}

	return // parent exists, app dir does not isNotExist return
}

// checks that appName is usable
func (d *AppDirectory) checkAppName() (appName string, err error) {

	if appName = d.App; appName == "" {
		err = perrors.NewPF("appName cannot be empty")
		return // empty error return
	}

	for i, c := range appName {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) {
			err = perrors.ErrorfPF(
				"appName can only contain Unicode letters or digits: #%d: %q",
				i, c,
			)
			return // bad character error return
		}
	}

	return // good return
}

// eval evaluates the full app directory path
//   - on success, updates d.abs
func (d *AppDirectory) eval(path string) (abs string, isNotExist bool, err error) {
	var a string

	if a, err = pfs.AbsEval(path); err != nil {
		isNotExist = punix.IsENOENT(err)
		return // some error including does not exist
	}

	// success, app directory exists and is evaled
	d.abs.CompareAndSwap(nil, &a)
	abs = a

	return // success, directory exists
}
