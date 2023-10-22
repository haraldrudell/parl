/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"bufio"
	"os"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Keystrokes reads line-wise from standard input
type Keystrokes struct {
	stdin parl.NBChan[string]
}

// NewKeystrokes returns an object reading lines from standard input
//   - -verbose='mains...Keystrokes|mains.keystrokesThread' “github.com/haraldrudell/parl/mains.(*Keystrokes)”
func NewKeystrokes() (keystrokes *Keystrokes) {
	return &Keystrokes{}
}

// Launch starts reading stdin for keystrokes
//   - keystrokesThread reads blocking from os.Stdin and therefore cannot be cancelled
//   - therefore no parl.Go is used
//   - parl.Infallible prints errors that should not happen to os.Stderr
//   - on CloseNow, KeystrokesThread exits on the following return keypress.
//     Mostly, the process exits prior.
func (k *Keystrokes) Launch() (k2 *Keystrokes) {
	k2 = k
	go keystrokesThread(&k.stdin)
	return
}

// Ch returns a channel where lines from the keyboard can be received
//   - the channel may close upon [Keystrokes.CloseNow]
func (k *Keystrokes) Ch() (ch <-chan string) {
	return k.stdin.Ch()
}

func (k *Keystrokes) CloseNow(errp *error) {
	k.stdin.CloseNow(errp)
}

// keystrokesThread reads blocking from os.Stdin therefore cannot be cancelled.
//   - therefore no parl.Go
//   - parl.Infallible prints errors that should not happen to os.Stderr
//   - on CloseNow, KeystrokesThread exits on the following return
//   - -verbose=mains.keystrokesThread
func keystrokesThread(stdin *parl.NBChan[string]) {
	var err error
	defer parl.Debug("keystrokes.scannerThread exiting: err: %s", perrors.Short(err))
	defer parl.Recover(parl.Annotation(), &err, parl.Infallible)

	var scanner = bufio.NewScanner(os.Stdin)
	parl.Debug("keystrokes.scannerThread scanning: stdin.Ch: 0x%x", stdin.Ch())
	for scanner.Scan() { // scanner is typically stuck here
		// DidClose is true if close was invoked.
		// stdin.Ch may not be closed yet.
		if stdin.DidClose() {
			return // terminated by Keystrokes.CloseNow
		}
		if parl.IsThisDebug() {
			parl.Log("keystrokes.Send %q", scanner.Text())
		}
		stdin.Send(scanner.Text())
	}

	// scanner had error
	err = scanner.Err()
}
