/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// EofReader is an empty reader returning EOF. Thread-safe
package pio

import "io"

type eofReader struct{}

// EofReader is an empty reader returning EOF. Thread-safe
var EofReader io.Reader = &eofReader{}

// Read always return io.EOF menaing end-of-file
func (eof *eofReader) Read(p []byte) (n int, err error) {
	err = io.EOF
	return
}
