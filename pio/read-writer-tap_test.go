/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"errors"
	"testing"
)

func TestReadWriterTap(t *testing.T) {
	//t.Errorf("logging on")
	var e = errors.New("e message")
	var _ PIOErrorSource
	var readsError = NewPioError(PeReads, e)
	var err error
	_ = 1
	err = readsError
	var pioe *PIOError
	t.Logf("err: %T %T %q %t", &err, err, err.Error(), errors.As(err, &pioe))
	_ = err
}
