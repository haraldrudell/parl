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
//   - silent true: nothing is printed on os.Stdin closing
//   - silent false: “mains.keystrokesThread standard input closed” may be printed to
//     standard error on os.Stdin closing
//   - errorSink receives any errors returned by or panic in [os.Stdin.Read]
//   - stdin receives text lines from standard input with line terminator removed
//   - on [Keystrokes.CloseNow], keystrokesThread exits on the following keypress
//   - on [os.Stdin] closing, keystrokesThread closes the stdin line-input channel
//   - [StdinReader] converts any error to [io.EOF]
//   - [parl.Infallible] prints any errors to standard error, should not be any
//   - —
//   - -verbose=mains.keystrokesThread
func keystrokesThread(silent bool, errorSink parl.ErrorSink1, stdin parl.ClosableSink[string]) {
	var err error
	var isDebug = parl.IsThisDebug()
	if isDebug {
		defer parl.Debug("keystrokes.scannerThread exiting: err: %s", perrors.Short(err))
	}
	if errorSink == nil {
		errorSink = parl.Infallible
	}
	// if a panic is recovered, or err holds an error, both are printed to standard error
	defer parl.Recover(func() parl.DA { return parl.A() }, &err, errorSink)
	// ensure string-channel closes on exit without discarding any input
	defer stdin.EmptyCh()

	var isStdinReaderError atomic.Bool
	// scanner splits input into lines
	var scanner = bufio.NewScanner(NewStdinReader(errorSink, &isStdinReaderError))
	var stdinClosedCh = stdin.EmptyCh(parl.CloseAwaiter)

	// blocks here
	for scanner.Scan() {

		// check if consumer closed stdin output
		select {
		case <-stdinClosedCh:
			return // terminated by Keystrokes.CloseNow
		default:
		}
		if isDebug {
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
		isStdinReaderError.Load() { //	- close is caused by an error handled by StdinReader
		return
	}
	// echoed to standard error
	//	- echoed if:
	//	- stdin closed without error, eg. from user pressing ^D
	//	- silent is false
	parl.Log("%s standard input closed", perrors.PackFunc())
}
