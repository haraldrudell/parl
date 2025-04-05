/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"bufio"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/mains/malib"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog/plogt"
	"github.com/haraldrudell/parl/pruntime"
)

// Keystrokes reads line-wise from standard input
//   - [os.File.Read] from [os.Stdin] cannot be aborted because
//     Stdin cannot be closed
//   - therefore, on process exit or [Keystrokes.CloseNow], keystrokesThread thread is left blocked in Read
//   - —
//   - -verbose='mains...Keystrokes|mains.keystrokesThread' “github.com/haraldrudell/parl/mains.(*Keystrokes)”
//   - this object is released on process exit.
//     Remaining items due to stdin cannot be closed are:
//   - the stdin unbound channel
//   - optional addError
//   - those items along with [KeyStrokesThread] and [StdinReader] are released
//     on process exit or next keypress
type Keystrokes struct {
	// unbound thread-safe queue using lock and closing-channel wait-mechanic
	stdin parl.AwaitableSlice[string]
	// stdinReader is fail-free reader of standard input
	stdinReader *malib.StdinReader
}

// NewKeystrokes returns an object reading lines from standard input
//   - [Keystrokes.Launch] launches a thread reading from [os.Stdin]
//   - [Keystrokes.StringSource] provides a channel sending strings on each return key-press
//   - [Keystrokes.CloseNow] closes the channel discarding buffered characters
//   - Ch also closes on Stdin closing or thread runtime error
//
// Usage:
//
//	var err error
//	…
//	var keystrokes = NewKeystrokes()
//	defer keystrokes.Launch().CloseNow(&err)
//	for line := range keystrokes.Ch() {
func NewKeystrokes(fieldp ...*Keystrokes) (keystrokes *Keystrokes) {

	// get keystrokes
	if len(fieldp) > 0 {
		keystrokes = fieldp[0]
	}
	if keystrokes == nil {
		keystrokes = &Keystrokes{}
	}

	*keystrokes = Keystrokes{}
	return
}

// Launch starts reading stdin for keystrokes
//   - errorSink:
//   - silent missing:
//   - can only be invoked once per process or panic
//   - supports functional chaining
//   - silent [SilentClose] does not echo anything on [os.Stdin] closing
//   - addError if present receives errors from [os.Stdin.Read]
func (k *Keystrokes) Launch(errorSink parl.ErrorSink1, silent ...SilentType) (keystrokes *Keystrokes) {
	keystrokes = k

	// ensure only launched once
	if !didLaunch.CompareAndSwap(false, true) {
		var err = perrors.ErrorfPF("invoked multiple times")
		parl.Log(err.Error())
		panic(err) // terminates the process
	}

	var isSilent bool
	if len(silent) > 0 {
		isSilent = silent[0] == SilentClose
	}

	go k.stdinReaderThread(isSilent, errorSink)

	return
}

// StringSource returns a possibly closing unbound slice
// sending lines from the keyboard on each return press
//   - closing-channel wait mechanic
//   - Ch sends strings with return character removed
//   - the channel closes upon:
//   - — [Keystrokes.CloseNow] or
//   - — [os.Stdin] closing or
//   - — thread runtime error
func (k *Keystrokes) StringSource() (stringSource parl.ClosableSource1[string]) { return &k.stdin }

// CloseNow closes the string-sending channel discarding any pending characters
func (k *Keystrokes) CloseNow(errp *error) { k.stdin.Close() }

// keystrokesThread reads blocking from [os.Stdin] therefore cannot be canceled
//   - silent true: nothing is printed on os.Stdin closing
//   - silent false: “mains.keystrokesThread standard input closed” may be printed to
//     standard error on os.Stdin closing
//   - errorSink present: receives any errors returned by or panic in [os.Stdin.Read]
//   - errorSink nil: errors are printed to standard error
//   - stdin receives text lines from standard input with line terminator removed
//   - on [os.Stdin] closing, keystrokesThread closes the stdin line-input channel
//   - — the close may be deferred until a key is pressed or the process exits
//     -
//   - Because [os.Stdin] cannot be closed and [os.Stdin.Read] is blocking:
//   - — the thread may blockindfinitiely until process exit
//   - — therefore, keystrokesThread is a top-level function not waited upon
//   - — purpose is to minimize objects kept in memory until the thread exits
//   - on [Keystrokes.CloseNow], keystrokesThread exits on the following keypress
//   - [StdinReader] converts any error to [io.EOF]
//   - [parl.Infallible] prints any errors to standard error, should not be any
//   - —
//   - -verbose=Keystrokes..stdinReaderThread
func (k *Keystrokes) stdinReaderThread(silent bool, errorSink parl.ErrorSink1) {
	var err error
	var isDebug = parl.IsThisDebug()
	if isDebug {
		defer plogt.D("keystrokes.stdinReaderThread exiting: err: “%s”", perrors.Short(err))
	}
	if errorSink == nil {
		errorSink = parl.Infallible
	}
	// if a panic is recovered, or err holds an error, both are printed to standard error
	defer parl.Recover(func() parl.DA { return parl.A() }, &err, errorSink)
	// ensure string-channel closes on exit without discarding any input
	defer k.stdin.Close()

	var isStdinReaderError atomic.Bool
	var s = malib.NewStdinReader(errorSink, &isStdinReaderError)
	k.stdinReader = s
	// scanner splits input into lines
	var scanner = bufio.NewScanner(s)
	var stdinClosedCh = k.stdin.CloseCh()

	if isDebug {
		plogt.D("keystrokes.stdinReaderThread at for")
	}

	// blocks here
	for scanner.Scan() {

		// check if consumer closed stdin output
		select {
		case <-stdinClosedCh:
			return // terminated by Keystrokes.CloseNow
		default:
		}
		if isDebug {
			plogt.D("keystrokes.Send %q", scanner.Text())
		}
		k.stdin.Send(scanner.Text())
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
	parl.Log("[%s standard input closed]", pruntime.PackFunc())
}

// didLaunch ensures multiple keystrokesThread are not running
var didLaunch atomic.Bool
