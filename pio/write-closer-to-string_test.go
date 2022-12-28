/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"testing"
)

func TestWriteCloserToString(t *testing.T) {
	//t.Fail()

	var writer io.WriteCloser
	var n int
	var err error

	writer = NewWriteCloserToString()
	if err = writer.Close(); err != nil {
		t.Errorf("NewWriteCloserToString.Close err: %v", err)
	}

	// writer.Write: n: 0 err: "pio.Write file alread closed
	n, err = writer.Write([]byte{1})
	if err == nil {
		t.Error("writer.Write err missing")
	}
	t.Logf("writer.Write: n: %d err: %q", n, err)
}
