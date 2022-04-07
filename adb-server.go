/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package adb defines ageneric interface do Android Debug Bridge
package parl

import (
	"context"
	"io"
)

/*
Server is a generic representation of an adb server running on a host.

As of Android 12, Android Debug Bridge version 1.0.41 Version 32.0.0-8006631 has the following commands are supported:
devices connect disconnect pair forward ppp reverse mdns push pull sync shell emu
install install-multiple install-multiple-package uninstall bugreport jdwp logcat disable-verity enable-verity keygen
wait-for* get-state get-serialno get-devpath remount reboot sideload
root unroot usb tcpip start-server kill-server reconnect attach detach
*/
type Server interface {
	// DeviceSerialList lists serials for the currently online Android devices
	DeviceSerialList() (serials []AndroidSerial, err error)
	// DeviceForSerial obtains AdbDevice for a serial.
	// The device instance can execute additional device functions.
	// Devicette can connect to the device similarly to how the server implementation
	// connects to trhe adb Android Debug Bridge server
	DeviceForSerial(serial AndroidSerial) (android Devicette, err error)
	// TrackDevices emits serial numbers for devices that come online.
	// serials are sent on the serials channel.
	// if err is non-nil, set-up of failed.
	// The errs channel closes when watching stops
	// Watching is stopped by calling cancel function or when the server’s context terminates
	TrackDevices() (tracker Tracker, err error)
	/*
		NIMP 220405:
		host:version host:kill host:devices-l host:emulator: host:transport-usb
		host:transport-local host:transport-any host-serial: host-usb: host-local:
		host: :get-product :get-serialno :get-devpath :get-state
		:forward: :forward:norebind: :killforward: :killforward-all :list-forward
	*/
}

// Tracker represents a server connection emitting device serial numbers
type Tracker interface {
	// Serials emit serial number as online devices become available
	Serials() (serials <-chan AndroidSerial)
	// Errs is available once Serials close. It returns any errors
	Errs() (err error)
	// Cancel shuts down the Tracker
	Cancel()
}

/*
Devicette respresents a minimal adb protocol level connection to an Android device
via an adb server.
*/
type Devicette interface {
	// Shell executes a shell command on the device.
	// The resulting socket can be obtained either using the reader callback,
	// which is a socket connection to the device,
	// or by collecting the out string.
	Shell(command string, reader func(conn io.ReadWriteCloser) (err error)) (out string, err error)
	// Serial returns the device serial number
	Serial() (serial AndroidSerial)
}

// ServeretteFactory is a Server connection factory for Adbette implementations
type ServeretteFactory interface {
	// Adb connects to an adb adb Android Debug Bridge server on a specified tcp socket.
	// address is a string default "localhost:5037" and default port ":5037".
	// adbetter is a factory for Adbette connections.
	Adb(address string, adbetter Adbetter, ctx context.Context) (server Server)
}

// ServerFactory describes how AdbServer objects are obtained.
// Such servers may use duifferent protocol implementations from Adbette
type ServerFactory interface {
	// Adb connects to an adb adb Android Debug Bridge server on a specified tcp socket
	Adb(address string, ctx context.Context) (server Server)
	// AdbLocalhost connects to an adb Android Debug Bridge server on the local computer
	AdbLocalhost(ctx context.Context) (server Server)
}

// AndroidSerial uniquely identities an Anroid device.
// It is typically a string of a dozen or so 8-bit chanacters consisting of
// lower and upper case a-zA-Z0-9
type AndroidSerial string

// AdressProvider retrieves the address from an adb server or device so that
// custom devices can be created
type AdressProvider interface {
	DialAddress() (address string)
}
