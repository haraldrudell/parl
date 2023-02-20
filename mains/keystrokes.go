/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"bufio"
	"os"

	"github.com/haraldrudell/parl"
)

type Keystrokes struct {
	stdin parl.NBChan[string]
}

func NewKeystrokes() (keystrokes *Keystrokes) {
	return &Keystrokes{}
}

// Launch starts reading stdin for keystrokes
//   - keystrokesThread reads blocking from os.Stdin and therefore cannot be cancelled
//   - therefore no parl.Go is used
//   - parl.Infallible prints errors that should not happen to os.Stderr
//   - on CloseNow, KeystrokesThread exits on the following return keypress.
//     Mostly, the process exits prior.
func (ks *Keystrokes) Launch() (k2 *Keystrokes) {
	k2 = ks
	go keystrokesThread(&ks.stdin)
	return
}

func (ks *Keystrokes) Ch() (ch <-chan string) {
	return ks.stdin.Ch()
}

func (ks *Keystrokes) CloseNow(errp *error) {
	ks.stdin.CloseNow(errp)
}

// keystrokesThread reads blocking from os.Stdin therefore cannot be cancelled.
//   - therefore no parl.Go
//   - parl.Infallible prints errors that should not happen to os.Stderr
//   - on CloseNow, KeystrokesThread exits on the following return
func keystrokesThread(stdin *parl.NBChan[string]) {
	var err error
	defer parl.Debug("keystrokes.scannerThread exiting")
	defer parl.Recover(parl.Annotation(), &err, parl.Infallible)

	scanner := bufio.NewScanner(os.Stdin)
	parl.Debug("keystrokes.scannerThread scanning")
	for scanner.Scan() { // scanner is typically stuck here
		if stdin.DidClose() {
			return // terminated by Keystrokes.CloseNow
		}
		stdin.Send(scanner.Text())
	}

	// scanner had error
	err = scanner.Err()
}
