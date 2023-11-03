/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"
	"testing"
)

func TestEndCallbacks(t *testing.T) {
	var err = ErrEndCallbacks
	if !errors.Is(err, ErrEndCallbacks) {
		t.Error("err not ErrEndCallbacks")
	}
	var err2 = errors.New("x")
	if errors.Is(err2, ErrEndCallbacks) {
		t.Error("err2 ErrEndCallbacks")
	}
}
