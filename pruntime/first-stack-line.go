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
	stackBufferSize   = 100
	onlyThisGoroutine = false
	newlineByte       = 10
)

// FirstStackLine efficiently obtains the first line of a [runtime.Stack]
//   - "goroutine 34 [running]:\n…"
func FirstStackLine() (firstStackLine string) {
	var buffer = make([]byte, stackBufferSize)
	runtime.Stack(buffer, onlyThisGoroutine)
	if index := bytes.IndexByte(buffer, newlineByte); index != -1 {
		buffer = buffer[:index]
	}
	firstStackLine = string(buffer)
	return
}
