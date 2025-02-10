/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"io"
)

// Adbette is a minimal implementation of the adb Android debug bridge protocol.
// Adbette include both adb server and Android device functions.
// Adbette is extensible in that additional protocol features are easily
// implemented without concern for protocol details.
// to shutdown an Adbette and release its resouces, invoke the Closer method
type Adbette interface {
	// SendReadOkay sends a request to a remote adb endpoint.
	// if anything else than OKAY is received back from the
	// remote endpoint, err is non-nil.
	SendReadOkay(request AdbRequest) (err error)
	// ReadString reads utf-8 text up to 64 KiB-1 in length
	ReadString() (s string, err error)
	// ConnectToDevice sends a forwarding request to an adb
	// server to connect to one of its devices
	ConnectToDevice(serial AndroidSerial) (err error)
	// Shell executes a shell command on a device connected to the adb server.
	//	- out is a combination of stderr and stdout.
	//	- The status code from an on-device command cannot be obtained
	Shell(command string) (out []byte, err error)
	// ShellStream executes a shell command on the device returning a readable socket
	ShellStream(command string) (conn io.ReadWriteCloser, err error)
	// TrackDevices orders a server to emit serial number as they become available
	TrackDevices() (err error)
	// Devices lists the currently online serials
	Devices() (serials []AndroidSerial, err error)
	// DeviceStati returns all available serials and their status
	// The two slices correspond and are of the same length
	DeviceStati() (serials []AndroidSerial, stati []AndroidStatus, err error)
	// Closer closes an adb connection, meant to be a deferred function
	Closer(errp *error)
	// SetSync sets the protocol mode of an adb device connection to sync
	SetSync() (err error)
	// LIST is a sync request that lists file system entries in a directory of an adb device
	LIST(remoteDir string, dentReceiver func(mode uint32, size uint32, time uint32, byts []byte) (err error)) (err error)
	// RECV fetches the contents of a file on an adb device
	RECV(remotePath string, blobReceiver func(data []byte) (err error)) (err error)
	// CancelError is a value that a LIST or RECV callback routines can return to
	// cancel further invocations
	CancelError() (cancelError error)

	// below are convenience methods for extending Adbetter

	SendBlob(syncRequest AdbSyncRequest, data []byte) (err error)
	ReadBlob() (byts []byte, err error)
	ReadResponseID() (responseID AdbResponseID, err error)
	ReadBytes(byts []byte) (err error)
	SendBytes(byts []byte) (err error)
}

/*
AdbSocketAddress is a tcp socket address accessible to the
local host.
The format is two parts separated by a colon.
The first part is an IP address or hostname.
The second part is a numeric port number.
The empty string "" represents "localhost:5037".
If the port part is missing, such as "localhost" it implies port 5037.
If the host part is missing, it implies "localhost".
Note that localhost is valid both for IPv4 and IPv6.
*/
type AdbSocketAddress string

// AndroidSerial uniquely identities an Android device.
//   - has .String and .IsValid, is Ordered
//   - typically a string of a dozen or so 8-bit characters consisting of
//     lower and upper case a-zA-Z0-9
type AndroidSerial string

// Adbetter is a factory instance for connections featuring Adbette
//   - context is used for terminating a connection attempt.
//   - context is not retained in the connection.
type Adbetter interface {
	NewConnection(address AdbSocketAddress, ctx context.Context) (conn Adbette, err error)
}

// AdbRequest is a string formatted as an adb request.
// AdbRequest is only required for implementations using the Adbette protocol
// impementation
type AdbRequest string

type AdbSyncRequest string

type AdbResponseID string

// AdressProvider retrieves the address from an adb server or device so that
// custom devices can be created
type AdbAdressProvider interface {
	// AdbSocketAddress retrievs the tcp socket address used
	// by a near Adbette implementation
	AdbSocketAddress() (socketAddress AdbSocketAddress)
}
