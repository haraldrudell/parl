/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// do not echo to standard error on [os.Stdin] closing
	// optional argument to [Keystrokes.Launch]
	SilentClose = true
)

// didLaunch ensures multiple keystrokesThread are not running
var didLaunch atomic.Bool

// Keystrokes reads line-wise from standard input
//   - [os.File.Read] from [os.Stdin] cannot be aborted because
//     Stdin cannot be closed
//   - therefore, on process exit or [Keystrokes.CloseNow], keystrokesThread thread is left blocked in Read
//   - —
//   - -verbose='mains...Keystrokes|mains.keystrokesThread' “github.com/haraldrudell/parl/mains.(*Keystrokes)”
type Keystrokes struct {
	// unbound locked combined channel and slice type
	stdin parl.NBChan[string]
}

// NewKeystrokes returns an object reading lines from standard input
//   - [Keystrokes.Launch] launches a thread reading from [os.Stdin]
//   - [Keystrokes.Ch] provides a channel sending strings on each return key-press
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
func NewKeystrokes() (keystrokes *Keystrokes) { return &Keystrokes{} }

// Launch starts reading stdin for keystrokes
//   - can only be invoked once per process or panic
//   - supports functional chaining
//   - silent [SilentClose] does not echo anything on [os.Stdin] closing
func (k *Keystrokes) Launch(addError parl.AddError, silent ...bool) (keystrokes *Keystrokes) {
	keystrokes = k

	// ensure only launched once
	if !didLaunch.CompareAndSwap(false, true) {
		var err = perrors.ErrorfPF("invoked multiple times")
		parl.Log(err.Error())
		panic(err) // terminates the process
	}

	var isSilent bool
	if len(silent) > 0 {
		isSilent = silent[0]
	}

	go keystrokesThread(isSilent, addError, &k.stdin)

	return
}

// Ch returns a possibly closing receive-only channel sending lines from the keyboard on each return press
//   - Ch sends strings with return character removed
//   - the channel closes upon:
//   - — [Keystrokes.CloseNow] or
//   - — [os.Stdin] closing or
//   - — thread runtime error
func (k *Keystrokes) Ch() (ch <-chan string) { return k.stdin.Ch() }

// CloseNow closes the string-sending channel discarding any pending characters
func (k *Keystrokes) CloseNow(errp *error) {
	k.stdin.CloseNow(errp)
}
