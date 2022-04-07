/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"io"
)

// Adbetter is a factory instance for connections featuring Adbette
type Adbetter interface {
	NewConnection(address string, ctx context.Context) (conn Adbette, err error)
}

// Adbette is a minimal implementation of the adb Android debug bridge protocol
// Adbette include both adb server and Android device functions
type Adbette interface {
	// SendReadOkay sends a command to a the adb server.
	// if anything else than OKAY is received back from the
	// server, err is non-nil.
	SendReadOkay(s string) (err error)
	// ReadString reads utf-8 text from an adb server or device up to 64 KiB-1 in length
	ReadString() (s string, err error)
	// ConnectToDevice sends a forwarding request to an adb
	// server to connect to one of its devices
	ConnectToDevice(serial AndroidSerial) (err error)
	// Shell executes a shell command on a device connected to the adb server
	Shell(command string, reader func(conn io.ReadWriteCloser) (err error)) (out string, err error)
	// TrackDevices orders a server to emit serial number as they become available
	TrackDevices() (err error)
	// Devices lists the currently online serials
	Devices() (serials []AndroidSerial, err error)
	// DeviceStati returns all available serials and their status
	// The two slices correspond and are of the same length
	DeviceStati() (serials []AndroidSerial, stati []AndroidStatus, err error)
	// Closer closes an adb connection, meant to be a deferred function
	Closer(errp *error)
	// SetSync chnages the protocol mode of an adb connection to sync
	SetSync() (err error)
	// LIST is a sync request that lists the entriues in a directory of an adb device
	LIST(remoteDir string, dentReceiver func(mode uint32, size uint32, time uint32, byts []byte) (err error)) (err error)
	// RECV fetches the contents of a file located on an adb device
	RECV(remotePath string, blobReceiver func(data []byte) (err error)) (err error)
	// CancelError is a value that a LIST or RECV callback routine can return to
	// cancel further invocations
	CancelError() (cancelError error)
}
