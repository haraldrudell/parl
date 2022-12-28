/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"testing"
)

func TestNewWriteCloserToChanLine(t *testing.T) {
	lines := []string{"\n", "a\n\n", "", "d", "\ne"}
	exp := []string{"", "a", "", "d", "e"}

	var writeCloser io.WriteCloser
	var impl *WriteCloserToChanLine
	length := len(lines)
	bytss := make([][]byte, length)
	for i, s := range lines {
		bytss[i] = []byte(s)
	}

	writeCloser = NewWriteCloserToChanLine()
	impl = writeCloser.(*WriteCloserToChanLine)

	for _, byts := range bytss {
		writeCloser.Write(byts)
	}

	writeCloser.Close()

	ch := impl.Ch()
	for i := 0; i < len(exp); i++ {
		s, ok := <-ch
		if !ok {
			break
		}
		if s != exp[i] {
			t.Errorf("line %d: %q exp %q", i, s, exp[i])
		}
	}
}
