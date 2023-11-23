/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"runtime"
)

// StackTrace returns [runtime.Stack] after allocating sufficient buffer
//   - if the entire stackTrace is converted to string and split: those substrings
//     will be interned part of the larger stackTrace string causing memory leak, ie.
//     If only a single character is kept, the entire block is kept.
//     This leads to megabytes of memory leaks
//   - StackTrace returns a byte slice for convert smaller indiviual parts
//     to string
//   - the stack trace contains spaces, newlines and tab characters for formatting
//   - the first line is status line
//   - each frame is then two lines:
//   - — a function line with argument values
//   - — a filename line beginning with a tab character and
//     a hexadecimal in-line byte offset
//   - the first line-pair is for the StackTrace function itself
//   - if the executing thread is a goroutine:
//   - — the final two lines is “created by,” ie. the location of the go statement and
//     what thread started the goroutine
//   - — the two preceding lines is the goroutine function
//   - the stack trace has a terminating newline
func StackTrace() (stackTrace []byte) {
	var buf []byte
	var n int
	for size := 1024; ; size *= 2 {
		buf = make([]byte, size)
		if n = runtime.Stack(buf, runtimeStackOnlyThisGoroutine); n >= size {
			buf = nil
			continue
		}
		break
	}
	stackTrace = buf[:n]
	return
}
