/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"bytes"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

func TestLineReader(t *testing.T) {
	line1 := "line1\n"
	line2 := "line2\n"

	var readWriteCloserSlice *ReadWriteCloserSlice
	var byts = []byte(line1 + line2)
	var n int
	var err error
	var lineReader *LineReader
	var isEOF bool

	readWriteCloserSlice = NewReadWriteCloserSlice()
	if n, err = readWriteCloserSlice.Write(byts); err != nil {
		t.Errorf("readWriteCloserSlice.Write err: %s", perrors.Short(err))
		t.FailNow()
	} else if n != len(byts) {
		t.Errorf("n %d exp %d", n, len(byts))
		t.FailNow()
	}
	if err = readWriteCloserSlice.Close(); err != nil {
		t.Errorf("readWriteCloserSlice.Close err: %s", perrors.Short(err))
		t.FailNow()
	}

	lineReader = NewLineReader(readWriteCloserSlice)

	if byts, isEOF, err = lineReader.ReadLine(byts); err != nil {
		t.Errorf("ReadLine1 err: %s", perrors.Short(err))
	}
	if isEOF {
		t.Error("isEOF1 true")
	}
	if string(byts) != line1 {
		t.Errorf("byts1: %q exp %q", string(byts), line1)
	}

	if byts, isEOF, err = lineReader.ReadLine(byts); err != nil {
		t.Errorf("ReadLine2 err: %s", perrors.Short(err))
	}
	if isEOF {
		t.Error("isEOF2 true")
	}
	if string(byts) != line2 {
		t.Errorf("byts2: %q exp %q", string(byts), line2)
	}

	if byts, isEOF, err = lineReader.ReadLine(byts); err != nil {
		t.Errorf("ReadLine3 err: %s", perrors.Short(err))
	}
	if !isEOF {
		t.Error("isEOF3 false")
	}
	if string(byts) != "" {
		t.Errorf("byts3: %q exp %q", string(byts), "")
	}

}

func TestLineReader4k(t *testing.T) {
	line1 := bytes.Repeat([]byte("a"), 4096)

	var readWriteCloserSlice *ReadWriteCloserSlice
	var n int
	var err error
	var lineReader *LineReader
	var byts []byte
	var isEOF bool

	readWriteCloserSlice = NewReadWriteCloserSlice()
	if n, err = readWriteCloserSlice.Write(line1); err != nil {
		t.Errorf("readWriteCloserSlice.Write err: %s", perrors.Short(err))
		t.FailNow()
	} else if n != len(line1) {
		t.Errorf("n %d exp %d", n, len(line1))
		t.FailNow()
	}
	if err = readWriteCloserSlice.Close(); err != nil {
		t.Errorf("readWriteCloserSlice.Close err: %s", perrors.Short(err))
	}

	lineReader = NewLineReader(readWriteCloserSlice)

	if byts, isEOF, err = lineReader.ReadLine(byts); err != nil {
		t.Errorf("ReadLine1 err: %s", perrors.Short(err))
	}
	if !isEOF {
		t.Error("isEOF1 false")
	}
	if slices.Compare(byts, line1) != 0 {
		t.Errorf("byts1: %d exp %d", len(byts), len(line1))
	}
}
