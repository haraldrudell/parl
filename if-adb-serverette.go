/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
)

/*
Serverette is a generic representation of an adb server running on a host.

command-line adb:
As of Android 12, Android Debug Bridge version 1.0.41 Version 32.0.0-8006631 has the following commands are supported:
devices connect disconnect pair forward ppp reverse mdns push pull sync shell emu
install install-multiple install-multiple-package uninstall bugreport jdwp logcat disable-verity enable-verity keygen
wait-for* get-state get-serialno get-devpath remount reboot sideload
root unroot usb tcpip start-server kill-server reconnect attach detach
*/
type Serverette interface {
	AdbAdressProvider // AdbSocketAddress()
	// DeviceSerialList lists serials for the currently online Android devices
	DeviceSerialList() (serials []AndroidSerial, err error)
	// DeviceStati lists all serials currently known to the server along with
	// their current status.
	// The two slices correspond and are of the same length
	DeviceStati() (serials []AndroidSerial, stati []AndroidStatus, err error)
	// TrackDevices emits serial numbers for devices that come online.
	// serials are sent on the serials channel.
	// if err is non-nil, set-up of failed.
	// The errs channel closes when watching stops
	// Watching is stopped by calling cancel function or when the server’s context terminates
	TrackDevices() (tracker Trackerette, err error)
	/*
		NIMP 220405:
		host:version host:kill host:devices-l host:emulator: host:transport-usb
		host:transport-local host:transport-any host-serial: host-usb: host-local:
		host: :get-product :get-serialno :get-devpath :get-state
		:forward: :forward:norebind: :killforward: :killforward-all :list-forward
	*/
}

// Trackerette represents a server connection emitting device serial numbers
type Trackerette interface {
	// Serials emit serial number as online devices become available
	Serials() (serials <-chan AndroidSerial)
	// Errs is available once Serials close. It returns any errors
	Errs() (err error)
	// Cancel shuts down the Tracker
	Cancel()
}

// ServeretteFactory is a Server connection factory for Adbette implementations
type ServeretteFactory interface {
	// Adb connects to an adb adb Android Debug Bridge server on a specified tcp socket.
	// address is a string default "localhost:5037" and default port ":5037".
	// adbetter is a factory for Adbette connections.
	NewServerette(address AdbSocketAddress, adbetter Adbetter, ctx context.Context) (server Serverette)
}

// ServerFactory describes how AdbServer objects are obtained.
// Such servers may use duifferent protocol implementations from Adbette
type ServerFactory interface {
	// Adb connects to an adb adb Android Debug Bridge server on a specified tcp socket
	Adb(address AdbSocketAddress, ctx context.Context) (server Serverette)
	// AdbLocalhost connects to an adb Android Debug Bridge server on the local computer
	AdbLocalhost(ctx context.Context) (server Serverette)
}

// AndroidStatus indicates the current status of a device
// known to a Server or Serverette
// it is a single word of ANSII-set characters
type AndroidStatus string

// AndroidOnline is the Android device status
// that indicates an online device
//   - can be checked using method [AndroidSerial.IsOnline]
const AndroidOnline AndroidStatus = "device"
