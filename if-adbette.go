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
//   - Adbette include both adb server and Android device functions
//   - Adbette is extensible in that additional protocol features are easily
//     implemented without concern for protocol details.
//   - to shutdown an Adbette and release its resouces, invoke the Closer method
type Adbette interface {
	// SendReadOkay sends a request to a remote adb endpoint
	//	- request:
	//	- err: “OKAY” was not received back from the remote endpoint
	SendReadOkay(request AdbRequest) (err error)
	// ReadString reads utf-8 text up to 64 KiB-1 in length from the remote endpoint
	//	- s:
	//	- err:
	ReadString() (s string, err error)
	// ConnectToDevice sends a forwarding request to an adb
	// server to connect to one of its devices
	//	- serial:
	//	- err:
	ConnectToDevice(serial AndroidSerial) (err error)
	// Shell executes a shell command on a device connected to the adb server
	//	- command:
	//	- out: a combination of standard error and standard output
	//	- err:
	//	- —
	//	- The status code from an on-device command cannot be obtained
	Shell(command string) (out []byte, err error)
	// ShellStream executes a shell command on the device returning a readable socket
	//	- command:
	//	- conn:
	//	- err:
	ShellStream(command string) (conn io.ReadWriteCloser, err error)
	// TrackDevices orders a server to emit serial number as they become available
	//	- err:
	//	- —
	//	- blocking…
	TrackDevices() (err error)
	// Devices lists the currently online serials
	//	- serials: typically Android serial number “8BQX1E63Z”
	//		May be peculiar references if non-usb tcp/ip connections
	//	- err:
	Devices() (serials []AndroidSerial, err error)
	// DeviceStati returns all available serials and their status
	//	- serial: typically Android serial number “8BQX1E63Z”
	//		May be peculiar references if non-usb tcp/ip connections
	//	- stati:
	//	- err:
	//	- —
	// The two slices correspond and are of the same length
	DeviceStati() (serials []AndroidSerial, stati []AndroidStatus, err error)
	// Closer closes an adb connection. Deferrable.
	//	- errp
	Closer(errp *error)
	// SetSync sets the protocol mode of an adb device connection to sync
	//	- err:
	SetSync() (err error)
	// LIST is a sync request that lists file system entries in a directory of an adb device
	//	- remoteDir:
	//	- dentReceiver
	//	- err:
	LIST(remoteDir string, listReceiver AdbListReceiver) (err error)
	// RECV fetches the contents of a file on an adb device
	RECV(remotePath string, blobReceiver AdbBlobReceiver) (err error)
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

type AdbListReceiver interface {
	// LISTReceiver receives untyped data from adb LIST command
	//	- mode:
	//	- size:
	//	- time:
	//	- byts:
	//	- err:
	ReceiveLISTEntry(mode uint32, size uint32, time uint32, byts []byte) (err error)
}

type AdbBlobReceiver interface {
	// ReceiveBlob
	//	- data:
	//	- err:
	ReceiveBlob(data []byte) (err error)
}

// AdbSocketAddress is a tcp socket address accessible to the
// local host.
//   - The format is two parts separated by a colon.
//   - The first part is an IP address or hostname.
//   - The second part is a numeric port number.
//   - The empty string "" represents "localhost:5037".
//   - If the port part is missing, such as "localhost" it implies port 5037.
//   - If the host part is missing, it implies "localhost".
//   - Note that localhost is valid both for IPv4 and IPv6.
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

// AdbRequest is an encoded string
//   - AdbRequest is only required for implementations using the Adbette protocol
//     impementation
//   - 4 hexadecimal bytes followed by text-bytes
type AdbRequest []byte

func (r AdbRequest) String() (s string) {
	if len(r) < 5 {
		return "adbRequest-uninitialized"
	}
	s = string(r[4:])
	return
}

type AdbSyncRequest string

// AdbResponseID is 4 8-byte characters response
// as 4-byte array
type AdbResponseID [4]byte

func (r AdbResponseID) String() (s string) { return string(r[:]) }

// AdressProvider retrieves the address from an adb server or device so that
// custom devices can be created
type AdbAdressProvider interface {
	// AdbSocketAddress retrievs the tcp socket address used
	// by a near Adbette implementation
	AdbSocketAddress() (socketAddress AdbSocketAddress)
}
