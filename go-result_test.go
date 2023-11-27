/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestGoResult(t *testing.T) {
	var text = "text"
	var err error
	_ = 1
	err = runGoroutine(text)
	if err == nil {
		t.Error("err missing")
	} else if message := err.Error(); message != text {
		t.Errorf("err message: %q exp %q", message, text)
	}
}

func runGoroutine(text string) (err error) {
	var g = NewGoResult()
	go goroutine(text, g)
	defer g.ReceiveError(&err)

	return
}

func goroutine(text string, g GoResult) {
	var err error
	defer g.SendError(&err)
	defer RecoverErr(func() DA { return A() }, &err)

	err = perrors.New(text)
}
