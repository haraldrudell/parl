//go:build !android && !illumos && !ios && !js && !plan9 && !wasip1 && !windows

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"fmt"
	"os"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

// “signal “segmentation fault” SIGSEGV 11 0xb”
//   - [os.Signal] is a signal type available on all platforms: interface { String() string; Signal() }
//   - a signal is an operating-system dependent small non-zero positive integer
//   - not all operating systems, eg. Windows, Android, or iOS, have signals
//   - os.Signal is implemented on most platforms by [unix.Signal] of underlying type int
//   - values of unix.Signal are platform-dependent constants, eg. [unix.SIGINT]
//   - a signal has a name, a single upper case word like “SIGINT”
//   - [unix.SignalName] returns the name of any valid unix.Signal value.
//     Invalid signals returns their integer number “-1”
//   - a signal has a description, a short lower-case sentence like “segmentation fault”
//   - [unix.Signal.String] or [os.Signal.String] returns the description for a signal value.
//     Invalid signals returns a numeric value “signal -1”
//   - the lowest signal value for all known platforms is 1 SIGHUP
//   - it is not possible to retrieve the number of the last signal for a platform.
//     Typically, the highest value is 31 SIGUSR2, less than 100, and there are no intermediate ununsed numbers
//   - [unix.Signal] is as of go1.22.3 not defined for operating systems:
//     android illumos ios js plan9 wasip1 windows
func SignalString(signal os.Signal) (s string) {

	// the [unix.Signal] value contained in signal
	var unixSignal unix.Signal
	var isUnixSignal bool
	unixSignal, isUnixSignal = signal.(unix.Signal)
	if !isUnixSignal {
		panic(perrors.ErrorfPF("Unknown signal implementation: %T", signal))
	}

	// negative sign for printing heaxadecimal value
	var minus string
	// the positive value for signal
	//	- must be int or %x wil print the hexadecimal interpretation of String()
	var signalPositive = int(unixSignal)
	if signalPositive < 0 {
		minus = "-"
		signalPositive = -signalPositive
	}

	// signal name “SIGSEGV”
	//	- short upper-case word beginning with “SIG”
	//	- empty for unknown signal
	var signalName = unix.SignalName(unixSignal)
	if signalName != "" {
		signalName = "\x20" + signalName
	}

	// signal description “segmentation fault”
	// [unix.Signal.String] returns signal description
	//	- short lower-case word or sentence
	//	- empty for unknown signal
	var signalDesc = signal.String()
	if signalDesc != "" {
		signalDesc = "\x20“" + signalDesc + "”"
	}

	s = fmt.Sprintf(
		"signal%s%s %d %s0x%x",
		signalDesc, signalName,
		signal,
		minus, signalPositive,
	)

	return
}
