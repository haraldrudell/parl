/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"runtime"
)

func StackString() (stack string) {
	var buf []byte
	var n int
	for size := 1024; ; size *= 2 {
		buf = make([]byte, size)
		if n = runtime.Stack(buf, onlyThisGoroutine); n >= size {
			buf = nil
			continue
		}
		break
	}
	stack = string(buf[:n])
	buf = nil
	return
}
