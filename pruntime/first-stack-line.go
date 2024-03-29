/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"bytes"
	"runtime"
)

const (
	stackBufferSize               = 100
	runtimeStackOnlyThisGoroutine = false
	newlineByte                   = 10
)

// FirstStackLine efficiently obtains the first line of a [runtime.Stack]
//   - "goroutine 34 [running]:\n…"
//   - interning the first line as a string will cost about 25 bytes
func FirstStackLine() (firstStackLine []byte) {

	// beginning of a stack trace in a small buffer
	var buffer = make([]byte, stackBufferSize)
	runtime.Stack(buffer, runtimeStackOnlyThisGoroutine)
	if index := bytes.IndexByte(buffer, newlineByte); index != -1 {
		buffer = buffer[:index]
	}

	// byte sequence of 25 characters or so
	//	- interning large strings is temporary memory leak
	firstStackLine = buffer

	return
}
