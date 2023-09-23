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
//     If only a single character is kept, the entire block is kept
//   - therefore, split stackTrace as a byte slice and convert smaller indiviual parts
//     to string
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
