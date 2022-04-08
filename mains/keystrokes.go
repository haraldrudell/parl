/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"bufio"
	"os"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/ev"
	"github.com/haraldrudell/parl/perrors"
)

type scannerResult struct {
	hasText bool
	text    string
	err     error
	panic   bool
}

type InputLine string

// Keystrokes emits keystroke events
func Keystrokes(ctx ev.Callee) {
	defer parl.Recover(parl.Annotation(), nil, func(err error) { ctx.Failure(err) })

	lines := make(chan *scannerResult)
	go readLines(lines)
	var err error
	for {
		select {
		case line := <-lines:
			if line != nil {
				if line.hasText {
					ctx.Send((*InputLine)(&line.text))
					continue
				}
				err = line.err
			} else {
				err = perrors.New("Unexpected close of lines channel")
			}
		case <-ctx.Done():
		}
		break
	}
	if err == nil {
		ctx.Success()
	} else {
		ctx.Failure(parl.Errorf("stdio Scanner: '%w'", err))
	}
}

func readLines(lines chan<- *scannerResult) {
	parl.Recover(parl.Annotation(), nil, func(e error) { lines <- &scannerResult{err: e, panic: true}; close(lines) })
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines <- &scannerResult{hasText: true, text: scanner.Text()}
	}
	lines <- &scannerResult{err: scanner.Err()}
	close(lines)
}
