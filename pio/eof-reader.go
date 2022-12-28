/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import "io"

type eofReader struct{}

// EofReader returns a reader at EOF. Thread-safe
var EofReader io.Reader = &eofReader{}

func (eof *eofReader) Read(p []byte) (n int, err error) {
	err = io.EOF
	return
}
