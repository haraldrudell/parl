/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"bytes"
	"runtime"
	"runtime/debug"
)

const (
	// the lead-in line contains “goroutine 18 [running]:”
	prLeadInLines = 1
	// stack frame of pruntime.DebugStack function
	prDebugStackStdFrame = 1
	// stack frame of debuig.Stack function
	prDebugStackFnFrame = 1
	//	- first line of frame describes source file
	//	- second line of frame describes package and function
	prLinesPerFrame = 2
	// creator lines are two lines that describe how a goroutine was launched:
	//	- “created by…”
	prCreatorLines = 2
	// newline as a byte
	byteNewline = byte('\n')
	// tab as a byte
	byteTab = byte('\t')
	// space as a byte
	byteSpace = byte('\x20')
)

// byte slice newline separator
var byteSliceNewline = []byte{byteNewline}

// byte slice tab
var byteSliceTab = []byte{byteTab}

// byte slice two spaces
var byteSliceTwoSpaces = []byte{byteSpace, byteSpace}

var _ = runtime.Stack

// DebugStack returns a string stack trace intended to be printed or when a full printable trace is desired
//   - top returned stack frame is caller of [pruntime.DebugStack]
//   - skipFrames allows for removing additional frames.
//   - differences from debug.Stack:
//   - tabs are replaced with two spaces
//   - Stack frames other than the callers of pruntime.DebugStack are removed
func DebugStack(skipFrames int) (stack string) {
	if skipFrames < 0 {
		skipFrames = 0
	}

	/*
		goroutine 18 [running]:
		runtime/debug.Stack()
			/opt/homebrew/Cellar/go/1.18/libexec/src/runtime/debug/stack.go:24 +0x68
		github.com/haraldrudell/parl/pruntime.NewStack()
		…
		created by testing.(*T).Run
			/opt/homebrew/Cellar/go/1.18/libexec/src/testing/testing.go:1486 +0x300
	*/

	// remove final newline, split into lines
	//	- hold off on converting to string to reduce interning memory leak
	//	- [][]byte
	var stackTraceByteLines = bytes.Split(bytes.TrimSuffix(debug.Stack(), byteSliceNewline), byteSliceNewline)

	// number of stack frames that are not header data or creator line
	var frameCount = (len(stackTraceByteLines)-prLeadInLines-prCreatorLines)/prLinesPerFrame -
		prDebugStackStdFrame - prDebugStackFnFrame

	// check skipFrames maximum value
	if skipFrames > frameCount {
		skipFrames = frameCount
	}

	// undesirable stack frames: debug.Stack, pruntime.DebugStack and skipFrames
	var skipLines = prLinesPerFrame * (prDebugStackStdFrame + prDebugStackFnFrame + skipFrames)

	// remove lines
	copy(stackTraceByteLines[prLeadInLines:], stackTraceByteLines[prLeadInLines+skipLines:])
	stackTraceByteLines = stackTraceByteLines[:len(stackTraceByteLines)-skipLines]

	// merge back together, replace tab with two spaces, make string
	stack = string(bytes.ReplaceAll(
		bytes.Join(stackTraceByteLines, byteSliceNewline),
		byteSliceTab,
		byteSliceTwoSpaces,
	))

	return
}
