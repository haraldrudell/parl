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

// Keystrokes emits keystroke events
func Keystrokes(lines chan<- string, g0 parl.Go) {
	var err error
	defer g0.Done(err)
	defer parl.Recover(parl.Annotation(), &err, parl.NoOnError)
	defer parl.CloserSend(lines, &err)

	// we don’t need two threads because os.Stdin read cannot be aborted
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() { // scanner is typically stuck here
		if g0.Context().Err() != nil {
			return // terminated via context
		}
		lines <- scanner.Text()
	}

	// scanner had error
	err = scanner.Err()
}
