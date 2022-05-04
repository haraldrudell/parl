/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"path"
)

type AppDirectory struct {
	App string
}

// NewAppDir returns an object providing an app-specific file system location with defered directory creation
func NewAppDir(appName string) (appd *AppDirectory) {
	return &AppDirectory{App: appName}
}

// AppDir gets an app-specific file system directory
func AppDir(appName string) (directory string) {
	directory = NewAppDir(appName).Directory()
	return
}

const (
	dotLocalDir = ".local"
	shareDir    = "share"
)

func (appd *AppDirectory) Directory() (directory string) {
	directory = HomeDir(path.Join(dotLocalDir, shareDir, appd.App))
	return
}
