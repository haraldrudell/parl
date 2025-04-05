/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

func TestLineReaderRead(t *testing.T) {
	const (
		abcdE = "abcd\ne\n"
		aByte = "a"
	)
	var (
		// pThreeBytes is 3-byte buffer
		pThreeBytes = make([]byte, 3)
		// expRead1 is “abc”
		expRead1 = string(abcdE[:len(pThreeBytes)])
		// expLength2 is 2
		expLength2 = 2
		// expRead2 is “d\n”
		expRead2 = abcdE[len(pThreeBytes) : len(pThreeBytes)+expLength2]
		// expLength3 is 2
		expLength3 = len(abcdE) - len(pThreeBytes) - expLength2
		// expRead3 “e\n”
		expRead3 = string(abcdE[len(pThreeBytes)+expLength2:])
		// expLength4 zero
		expLength4 = 0
		expALength = 1
	)

	var (
		n       int
		err     error
		actualS string
	)

	// Read() ReadLine()
	var lineReader *LineReader
	_ = lineReader

	// Read should fill p completely
	lineReader = NewLineReader(bytes.NewReader([]byte(abcdE)))
	n, err = lineReader.Read(pThreeBytes)
	if err != nil {
		t.Errorf("FAIL Read err: %s", err)
	}
	if n != len(pThreeBytes) {
		t.Errorf("FAIL Read bad n: %d exp %d", n, len(pThreeBytes))
	}
	actualS = string(pThreeBytes)
	if actualS != expRead1 {
		t.Errorf("FAIL Read bad data: %q exp %q", actualS, expRead1)
	}

	// Read should stop at newline
	n, err = lineReader.Read(pThreeBytes)
	if err != nil {
		t.Errorf("FAIL Read err: %s", err)
	}
	if n != expLength2 {
		t.Errorf("FAIL Read bad n: %d exp %d", n, expLength2)
	} else {
		actualS = string(pThreeBytes[:n])
		if actualS != expRead2 {
			t.Errorf("FAIL Read bad data: %q exp %q", actualS, expRead2)
		}
	}

	// Read should return last line just before EOF
	n, err = lineReader.Read(pThreeBytes)
	if err != nil {
		t.Errorf("FAIL Read err: %s", err)
	}
	if n != expLength3 {
		t.Errorf("FAIL Read bad n: %d exp %d", n, expLength3)
	} else {
		actualS = string(pThreeBytes[:n])
		if actualS != expRead3 {
			t.Errorf("FAIL Read bad data: %q exp %q", actualS, expRead3)
		}
	}

	// Read should return EOF
	n, err = lineReader.Read(pThreeBytes)
	if err == nil {
		t.Error("FAIL missing error")
	} else if !errors.Is(err, io.EOF) {
		t.Errorf("FAIL Read bad err: %s", err)
	}
	if n != expLength4 {
		t.Errorf("FAIL Read bad n: %d exp %d", n, expLength4)
	}

	// Read no end of line EOF
	lineReader = NewLineReader(bytes.NewReader([]byte(aByte)))
	n, err = lineReader.Read(pThreeBytes)
	if err == nil {
		t.Error("FAIL missing error")
	} else if !errors.Is(err, io.EOF) {
		t.Errorf("FAIL Read bad err: %s", err)
	}
	if n != expALength {
		t.Errorf("FAIL Read bad n: %d exp %d", n, expALength)
	} else {
		actualS = string(pThreeBytes[:n])
		if actualS != aByte {
			t.Errorf("FAIL Read bad data: %q exp %q", actualS, expRead3)
		}
	}

	// Read should handle stream error
	lineReader = NewLineReader(newErrorStream())
	n, err = lineReader.Read(pThreeBytes)
	if err == nil {
		t.Error("FAIL missing error")
	} else if !errors.Is(err, os.ErrClosed) {
		t.Errorf("FAIL Read bad err: %s", err)
	}
	if n != 0 {
		t.Errorf("FAIL Read bad n: %d exp %d", n, 0)
	}
}

func TestLineReader(t *testing.T) {
	var (
		// pThreeBytes is 3-byte buffer
		pThreeBytes = make([]byte, 3)
	)

	var (
		n     int
		err   error
		line  []byte
		isEOF bool
	)

	// Read() ReadLine()
	var lineReader *LineReader
	_ = lineReader

	// Read should handle error
	lineReader = NewLineReader(newErrorStream())
	n, err = lineReader.Read(pThreeBytes)
	if err == nil {
		t.Error("FAIL missing error")
	} else if !errors.Is(err, os.ErrClosed) {
		t.Errorf("FAIL Read bad err: %s", err)
	}
	if n != 0 {
		t.Errorf("FAIL Read bad n: %d exp %d", n, 0)
	}

	// ReadLine should handle error
	lineReader = NewLineReader(newErrorStream())
	line, isEOF, err = lineReader.ReadLine()
	if err == nil {
		t.Error("FAIL missing error")
	} else if !errors.Is(err, os.ErrClosed) {
		t.Errorf("FAIL Read bad err: %s", err)
	}
	if isEOF {
		t.Error("FAIL isEOF1 true")
	}
	if line != nil {
		t.Error("FAIL line not nil")
	}
}

func TestLineReaderNoAlloc(t *testing.T) {
	const (
		line1       = "line1\n"
		line2       = "line2\n"
		unusualByte = byte('x')
	)
	var (
		// aLine “a\n”
		aLine = "a\n"
		// pThreeBytes is 3-byte buffer
		pThreeBytes = make([]byte, 3)
	)

	var (
		err     error
		isEOF   bool
		actualS string
		line    []byte
	)

	// Read() ReadLine()
	var lineReader *LineReader
	_ = lineReader

	// ReadLine should use p if line fits
	lineReader = NewLineReader(bytes.NewReader([]byte(aLine)))
	t.Logf("lineReader: %q", aLine)
	line, isEOF, err = lineReader.ReadLine(pThreeBytes)
	t.Logf("line: %q isEOF: %t err: %s", line, isEOF, perrors.Short(err))
	if err != nil {
		t.Errorf("FAIL ReadLine err: %s", err)
	}
	if isEOF {
		t.Error("FAIL isEOF1 true")
	}
	actualS = string(line)
	if actualS != aLine {
		t.Errorf("FAIL bad line: %q exp %q", actualS, aLine)
	}
	line[0] = unusualByte
	if line[0] != pThreeBytes[0] {
		t.Error("FAIL line unexpectedly reallocated")
	}

	// ReadLine should handle EOF after newline
	line, isEOF, err = lineReader.ReadLine(pThreeBytes)
	t.Logf("line: %q isEOF: %t err: %s", line, isEOF, perrors.Short(err))
	if err != nil {
		t.Errorf("FAIL Read err: %s", err)
	}
	if !isEOF {
		t.Error("FAIL isEOF false")
	}
	if line != nil {
		t.Error("FAIL line not nil on EOF")
	}
}

func TestLineReaderAlloc(t *testing.T) {
	const (
		line1       = "line1"
		unusualByte = byte('x')
	)
	var (
		// pThreeBytes is 3-byte buffer
		pThreeBytes = make([]byte, 3)
	)

	var (
		err     error
		isEOF   bool
		actualS string
		line    []byte
	)

	// Read() ReadLine()
	var lineReader *LineReader
	var reset = func() {
		lineReader = NewLineReader(bytes.NewReader([]byte(line1)))
	}

	// ReadLine should use allocated line when p is too small
	reset()
	line, isEOF, err = lineReader.ReadLine(pThreeBytes)
	if err != nil {
		t.Errorf("FAIL ReadLine err: %s", err)
	}
	if !isEOF {
		t.Error("FAIL isEOF false")
	}
	actualS = string(line)
	if actualS != line1 {
		t.Errorf("FAIL bad line: %q exp %q", actualS, line1)
	}
	line[0] = unusualByte
	if line[0] == pThreeBytes[0] {
		t.Error("FAIL line not reallocated")
	}

	// ReadLine should use allocated line when p missing
	reset()
	line, isEOF, err = lineReader.ReadLine()
	if err != nil {
		t.Errorf("FAIL Read err: %s", err)
	}
	if !isEOF {
		t.Error("FAIL isEOF false")
	}
	if line == nil {
		t.Error("FAIL line nil")
	} else {
		actualS = string(line)
		if actualS != line1 {
			t.Errorf("FAIL bad line: %q exp %q", actualS, line1)
		}
	}
}

func TestLineReader4k(t *testing.T) {
	// 4 KiB should reallocate r.b
	// 4 KiB and large p should not cause allocation

	const (
		// 4 KiB + 1
		length      = defaultAllocation + 1
		unusualByte = byte('x')
	)
	var (
		// 4 KiB ‘a’
		line1       = bytes.Repeat([]byte("a"), length)
		largeBuffer = make([]byte, length+1)
	)

	var (
		err        error
		line, b0   []byte
		isEOF      bool
		actualByte byte
	)

	// Read() ReadLine()
	var lineReader *LineReader
	var reset = func() {
		lineReader = NewLineReader(bytes.NewReader(line1))
	}

	// ReadLine 4K with large p should not reallocate line or readLine.B
	reset()
	b0 = lineReader.b
	if b0 == nil {
		t.Fatal("FAIL lineReader.b not allocated")
	}
	line, isEOF, err = lineReader.ReadLine(largeBuffer)
	if err != nil {
		t.Errorf("FAIL ReadLine err: %s", perrors.Short(err))
	}
	if !isEOF {
		t.Error("FAIL isEOF false")
	}
	if line == nil {
		t.Fatal("line nil")
	}
	if slices.Compare(line, line1) != 0 {
		t.Errorf("FAIL bad line: %d exp %d", len(line), len(line1))
	}
	if lineReader.b == nil {
		t.Fatal("lineReader.b not allocated")
	}
	line[0] = unusualByte
	if largeBuffer[0] != unusualByte {
		t.Error("line not largeBuffer but reallocated")
	}
	b0[0] = unusualByte
	actualByte = lineReader.b[0]
	if actualByte != unusualByte {
		t.Error("readLine.b was reallocated")
	}

	// ReadLine 4K should reallocate readLine.B for performance
	reset()
	b0 = lineReader.b
	if b0 == nil {
		t.Fatal("FAIL lineReader.b not allocated")
	}
	line, isEOF, err = lineReader.ReadLine()
	if err != nil {
		t.Errorf("FAIL ReadLine err: %s", perrors.Short(err))
	}
	if !isEOF {
		t.Error("FAIL isEOF false")
	}
	if line == nil {
		t.Fatal("line nil")
	}
	if slices.Compare(line, line1) != 0 {
		t.Errorf("FAIL bad line: %d exp %d", len(line), len(line1))
	}
	if lineReader.b == nil {
		t.Fatal("lineReader.b not allocated")
	}
	b0[0] = unusualByte
	actualByte = lineReader.b[0]
	if actualByte == unusualByte {
		t.Error("readLine.b was not reallocated")
	}
}

// errorStream is a stream returning os.ErrClosed
type errorStream struct{}

var _ io.Reader = &errorStream{}

// newErrorStream returns a stream returning os.ErrClosed
func newErrorStream() (e *errorStream) { return &errorStream{} }

// Read returns os.ErrClosed
func (e *errorStream) Read(p []byte) (n int, err error) {
	err = os.ErrClosed
	return
}
