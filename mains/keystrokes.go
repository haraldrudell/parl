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
	"github.com/haraldrudell/parl/perrors"
)

// Keystrokes emits keystroke events
// on g0.Context() shutdown, scannerThread is left running until the next newline
func Keystrokes(send func(s string), g0 parl.Go) {
	var err error
	defer g0.Done(&err)
	defer perrors.Deferr("Keystrokes exit", &err, parl.GetDebug(0))
	defer parl.Recover(parl.Annotation(), &err, parl.NoOnError)

	// stdio.Scan cannot be terminated, so let that thread terminate whenever
	// scanLines will be left open
	var scanLines parl.NBChan[string]
	// no errors are collected from scannerThread nor is it waited on
	go scannerThread(scanLines.Send, g0.Context())

	// consume scannerThread output
	parl.Debug("keystrokes at for")
	scannerCh := scanLines.Ch()
	done := g0.Context().Done()
	for {
		select {
		case line := <-scannerCh:
			send(line)
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
	defer parl.Debug("keystrokes.scannerThread exiting")
	defer parl.Recover(parl.Annotation(), &err, parl.Infallible)

	scanner := bufio.NewScanner(os.Stdin)
	parl.Debug("keystrokes.scannerThread scanning")
	for scanner.Scan() { // scanner is typically stuck here
		if ctx.Err() != nil {
			return // terminated via context
		}
		send(scanner.Text())
	}

	// scanner had error
	err = scanner.Err()
}
