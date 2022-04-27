/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"bufio"
	"context"
	"os"

	"github.com/haraldrudell/parl"
)

// Keystrokes emits keystroke events
// on g0.Context() shutdown, scannerThread is left running until the next newline
// the lines channel is never closed
func Keystrokes(lines chan<- string, g0 parl.Go) {
	var err error
	defer g0.Done(err)
	defer parl.Recover(parl.Annotation(), &err, parl.NoOnError)

	// stdio.Scan cannot be terminated, so let that thread terminate whenever
	var scanLines parl.NBChan[string]
	go scannerThread(scanLines.Send, g0.Context())

	// consume scannerThread output
	scannerCh := scanLines.Ch()
	done := g0.Context().Done()
	for {
		select {
		case line := <-scannerCh:
			lines <- line
		case <-done:
			return // canceled by context exit
		}
	}
}

// scannerThread reads from os.Stdin and therefore cannot be cancelled.
// send is a non-blocking send function.
// ctx indicates shutdown effective on next os.Stdin newline.
func scannerThread(send func(string), ctx context.Context) {
	var err error
	defer parl.Recover(parl.Annotation(), &err, parl.Infallible)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() { // scanner is typically stuck here
		if ctx.Err() != nil {
			return // terminated via context
		}
		send(scanner.Text())
	}

	// scanner had error
	err = scanner.Err()
}
