/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package phttp

import (
	"fmt"
	"log"
	"testing"
)

func TestLogCapturer_Write(t *testing.T) {
	const (
		value     = "value"
		expLength = 1
		expOutput = value + "\n"
	)
	var (
		printfer = newPrintfer()
	)

	// Printf is the method to test
	var logCapturer *log.Logger = NewLogCapturer(printfer.log)

	// Printf should log transparently
	logCapturer.Printf("%s", value)
	if len(printfer.output) != expLength {
		t.Fatalf("FAIL printfer.output len %d exp %d", len(printfer.output), expLength)
	}
	if s := printfer.output[0]; s != expOutput {
		t.Errorf("output %q exp %q", s, expOutput)
	}
}

type printfer struct {
	output []string
}

func newPrintfer() (p *printfer) { return &printfer{} }

func (p *printfer) log(format string, a ...any) {
	var s string
	if len(a) > 0 {
		s = fmt.Sprintf(format, a...)
	} else {
		s = format
	}
	p.output = append(p.output, s)
}
