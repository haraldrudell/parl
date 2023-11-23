/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"path"
)

// AppDirectory manages a per-user writable app-specific directory
type AppDirectory struct {
	App string // app name like “myapp”
}

// NewAppDir returns an object providing an app-specific file system location with defered directory creation
func NewAppDir(appName string) (appd *AppDirectory) {
	return &AppDirectory{App: appName}
}

// AppDir gets an app-specific writable file-system directory
//   - “~/.local/share/appName”
//   - ensures the directory exists with user-only permissions
func AppDir(appName string) (directory string) {
	directory = NewAppDir(appName).Directory()
	return
}

// Path returns the path to the directory without panics or file-system writes
func (appd *AppDirectory) Path() (directory string, err error) {
	var homeDir string
	if homeDir, err = UserHome(); err != nil {
		return
	}
	directory = path.Join(homeDir, dotLocalDir, shareDir, appd.App)

	return
}

const (
	// “.local” is a standardized directory name in a user’s home directory
	// on Linux
	dotLocalDir = ".local"
	// “share” is a standardized directory name in a user’s “~/.local” directory
	// on Linux
	shareDir = "share"
)

// Directory returns the app’s writable directory
//   - “~/.local/share/appName”
//   - ensures the directory exists with user-only permissions
func (appd *AppDirectory) Directory() (directory string) {
	directory = HomeDir(path.Join(dotLocalDir, shareDir, appd.App))
	return
}
