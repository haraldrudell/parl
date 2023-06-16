/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"runtime/debug"
	"strings"
)

/*
A stack has a leading line with go routine ID, then two lines per frame:
goroutine 1 [running]:
runtime/debug.Stack(0x1, 0x1, 0x2)

	/usr/local/Cellar/go/1.16.6/libexec/src/runtime/debug/stack.go:24 +0x9f
*/
const (
	// the number of lines debug.Stack produces for each stack frame
	linesPerStackFrame = 2
	// skip debug.Stack, that includes itself, and the Invocation stack frames
	skipFrames = 2
)

// Invocation returns an invocation stack trace for debug printing, empty string on troubles.
// The result is similar to the output from debug.Stack, but has some stack frames removed.
// tabs are replaced by two spaces.
// stackFramesToSkip 0 means first frame will be the caller of Invocation
// "goroutine 1 [running]:\ngithub.com/haraldrudell/parl/mains.(*Executable).AddErr(0x1809300, 0x158b620, 0xc000183800, 0x1) mains.(*Executable).AddErr-executable.go:302…"
func Invocation(stackFramesToSkip int) (stackTrace string) {
	if stackFramesToSkip < 0 {
		stackFramesToSkip = 0
	}

	// remove the first few stack frames
	stackBytes := debug.Stack()
	stackString := string(stackBytes)
	stackTraceLines := strings.Split(stackString, "\n")
	linesToSkip := (stackFramesToSkip + skipFrames) * linesPerStackFrame
	copy(stackTraceLines[1:], stackTraceLines[1+linesToSkip:])
	stackTraceLines = stackTraceLines[:len(stackTraceLines)-linesToSkip]
	stackTrace = strings.Join(stackTraceLines, "\n")

	return strings.ReplaceAll(stackTrace, "\t", "\x20\x20")
}
