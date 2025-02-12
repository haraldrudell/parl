/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/mains/malib"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// do not echo to standard error on [os.Stdin] closing
	// optional argument to [Keystrokes.Launch]
	SilentClose SilentType = true
)

// optional flag for no stdin-cose echo [SilentClose]
type SilentType bool

// didLaunch ensures multiple keystrokesThread are not running
var didLaunch atomic.Bool

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
	// unbound locked thread-safe slice with closing-channel wait mechanic
	stdin parl.AwaitableSlice[string]
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
	if len(fieldp) > 0 {
		keystrokes = fieldp[0]
	}
	if keystrokes != nil {
		*keystrokes = Keystrokes{}
	} else {
		keystrokes = &Keystrokes{}
	}
	return
}

// Launch starts reading stdin for keystrokes
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

	go malib.KeystrokesThread(isSilent, errorSink, &k.stdin)

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
func (k *Keystrokes) CloseNow(errp *error) { k.stdin.EmptyCh() }
