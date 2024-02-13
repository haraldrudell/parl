/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntimelib

import (
	"bytes"
)

// created by begins lines returned by runtime.Stack if
// the executing thread is a launched goroutine “created by ”
var runtCreatedByPrefix = []byte("created by ")

// ParseCreatedLine parses the second-to-last line of the stack trace.
// samples:
//   - “created by main.main”
//   - “created by main.(*MyType).goroutine1”
//   - “main.main()”
//   - go1.21.5 231219: “created by codeberg.org/haraldrudell/tools/gact.(*Transcriber).TranscriberThread in goroutine 9”
//   - no allocations
func ParseCreatedLine(createdLine []byte) (funcName, goroutineRef string, IsMainThread bool) {

	// remove prefix “created by ”
	var remain = bytes.TrimPrefix(createdLine, runtCreatedByPrefix)

	// if the line did not have this prefix, it is the main thread that launched main.main
	if IsMainThread = len(remain) == len(createdLine); IsMainThread {
		return // main thread: IsMainThread: true funcName zero-value
	}

	// “codeberg.org/haraldrudell/tools/gact.(*Transcriber).TranscriberThread in goroutine 9”
	if index := bytes.IndexByte(remain, '\x20'); index != -1 {
		funcName = string(remain[:index])
		goroutineRef = string(remain[index+1:])
	} else {
		funcName = string(remain)
	}

	return
}
