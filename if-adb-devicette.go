/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"io"
	"io/fs"
	"time"
)

// Devicette is a generic implementation of the capabilities
// of a device implementing the adb Android debug bridge protocol
type Devicette interface {
	// Serial returns the serial number for this device
	Serial() (serial AndroidSerial)
	// Shell executes a shell command on the device.
	//	- the response is a byte sequence
	//	- note: interning large strings may lead to memory leaks
	Shell(command string) (out []byte, err error)
	// ShellStream executes a shell command on the device returning a readable socket
	ShellStream(command string) (conn io.ReadWriteCloser, err error)
	/*
		Pull copies a remote file or directory on the Android device to a local file system location.
		the local file must not exist.
		Pull refuses certain files like product apks. shell cat works
	*/
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
		remount: dev: tcp: local:localreserved: localabstract: localfilesystem:
		framebuffer: jdwp: track-jdwp reverse:
		sync STAT SEND
	*/
}

// Dent is the information returned by adb ls or LIST
type Dent interface {
	// Name is utf-8 base path in device file system.
	// Name is base name, ie. file name and extension.
	Name() (name string)
	// Modified time, the time file contents was changed, second precision, continuous time
	Modified() (modified time.Time)
	// IsDir indicates directory.
	// LIST only support symbolic link, directory and regular file types
	IsDir() (isDir bool)
	// IsRegular indicates regular file, ie. not a directory or symbolic link.
	// LIST only support symbolic link, directory and regular file types
	IsRegular() (isRegular bool) // ie.not directory or symlink
	// Perm returns os.FileMode data.
	// 9-bit Unix permissions per os.FilePerm.
	// LIST also supports directory and symlink bits
	Perm() (perm fs.FileMode)
	// Size is limited to 4 GiB-1
	Size() (size uint32)
}

type AdbDentReceiver interface {
	ReceiveDent(dent Dent) (err error)
}

// DevicetteFactory describes how Devicette objects are obtained.
type DevicetteFactory interface {
	// NewDevicette creates a Devicette interacting with remote adb Android Debug Bridge
	// devices via an adb server available at the socket address address
	NewDevicette(address AdbSocketAddress, serial AndroidSerial, ctx context.Context) (devicette Devicette)
}
