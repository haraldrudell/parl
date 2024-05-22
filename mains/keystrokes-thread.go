/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"bufio"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// keystrokesThread reads blocking from [os.Stdin] therefore cannot be canceled
//   - therefore, keystrokesThread is a top-level function not waited upon
//   - on [Keystrokes.CloseNow], keystrokesThread exits on the following return keypress
//   - on [os.Stdin] closing, keystrokesThread closes the Keystrokes channel
//   - [StdinReader] converts any error to [io.EOF]
//   - [parl.Infallible] prints any errors to standard error, should not be any
//   - —
//   - -verbose=mains.keystrokesThread
func keystrokesThread(silent bool, errorSink parl.ErrorSink, stdin *parl.NBChan[string]) {
	var err error
	defer parl.Debug("keystrokes.scannerThread exiting: err: %s", perrors.Short(err))
	// if a panic is recovered, or err holds an error, both are printed to standard error
	defer parl.Recover(func() parl.DA { return parl.A() }, &err, parl.Infallible)
	// ensure string-channel closes on exit without discarding any input
	defer stdin.Close()

	var isError atomic.Bool
	var scanner = bufio.NewScanner(NewStdinReader(errorSink.AddError, &isError))
	parl.Debug("keystrokes.scannerThread scanning: stdin.Ch: 0x%x", stdin.Ch())

	// blocks here
	for scanner.Scan() {
		// DidClose is true if close was invoked
		//	- stdin.Ch may not be closed yet
		if stdin.DidClose() {
			return // terminated by Keystrokes.CloseNow
		}
		if parl.IsThisDebug() {
			parl.Log("keystrokes.Send %q", scanner.Text())
		}
		stdin.Send(scanner.Text())
	}

	// scanner had end of input or error
	//	- caused by a closing event like user pressing ^D or by
	//	- error during os.Stdin.Read or
	//	- error raised in Scanner
	err = scanner.Err()
	// do not print:
	if silent || //	- if silent is configured or
		err != nil || //	- the scanner had error or
		isError.Load() { //	- close is caused by an error handled by StdinReader
		return
	}
	// echoed to standard error
	//	- echoed if:
	//	- stdin closed without error, eg. from user pressing ^D
	//	- silent is false
	parl.Log("%s standard input closed", perrors.PackFunc())
}
