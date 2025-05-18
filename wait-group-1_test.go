/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
)

func TestWaitGroup1(t *testing.T) {
	//t.Error("Loggin on")
	var (
		isWinner, isDone bool
	)

	// AddWin() Done() IsDone()
	var w WaitGroup1

	// IsDone before Done is false
	isDone = w.IsDone()
	if isDone {
		t.Error("isDone true")
	}

	// first AddWin isWinner true
	isWinner = w.AddWin()
	if !isWinner {
		t.Error("isWinner false")
	}

	// Done after AddWin does not panic
	w.Done()

	// IsDone after Done is true
	isDone = w.IsDone()
	if !isDone {
		t.Error("isDone false")
	}

	// second AddWin isWinner false and does not hang
	isWinner = w.AddWin()
	if isWinner {
		t.Error("isWinner true")
	}
}
