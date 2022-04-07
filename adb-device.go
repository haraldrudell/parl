/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package adb defines ageneric interface do Android Debug Bridge
package parl

import (
	"io/fs"
	"time"
)

// Device is a generic representation of an Android Device accessible via an AdbServer
type Device interface {
	Devicette // Shell() Serial()
	// Pull copies a remote file or directory on the Android device to a local file system location.
	// the local file must not exist.
	Pull(remotePath, nearPath string) (err error)
	/*
		List has some peculiarities:
		If remoteDir is not an existing directory, an empty list is returned.
		Entries with insufficient permisions are ignored.
		Update: . and .. are removed, adb LIST ortherwise do return those.
		File mode: the only present type bits beyond 9-bit Unix permissions are
		symlink, regular file and directory.
		File size is limited to 4 GiB-1.
		Modification time resolution is second and range is confined to a 32-bit Unix timestamp.
	*/
	List(remoteDir string) (dFileInfo []Dent, err error)
	/*
		NIMP 220405:
		shell: remount: dev: tcp: local:localreserved: localabstract: localfilesystem:
		framebuffer: jdwp: track-jdwp reverse:
		sync STAT SEND
	*/
}

// Dent is the information returned by adb ls or LIST
type Dent interface {
	Name() (name string)            // utf-8
	Modified() (modified time.Time) // second precision, local time zone
	IsDir() (isDir bool)
	IsRegular() (isRegular bool) // ie.not directory or symlink
	Perm() (perm fs.FileMode)    // 9 bits Unix permissions, directory and symlink
	Size() (size uint32)         // limited to 4 GiB-1
}
