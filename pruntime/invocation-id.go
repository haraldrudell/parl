/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"errors"
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
	// debug.Stack uses this prefix in the first line of the result
	runtGoroutinePrefix = "goroutine "
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
	stackTraceLines := strings.Split(string(debug.Stack()), "\n")
	linesToSkip := (stackFramesToSkip + skipFrames) * linesPerStackFrame
	copy(stackTraceLines[1:], stackTraceLines[1+linesToSkip:])
	stackTraceLines = stackTraceLines[:len(stackTraceLines)-linesToSkip]
	stackTrace = strings.Join(stackTraceLines, "\n")

	return strings.ReplaceAll(stackTrace, "\t", "\x20\x20")
}

// GoRoutineID obtains a numeric string that as of Go1.18 is
// assigned to each goroutine. This number is an increasing
// unsigned integer beginning at 1 for the main invocation
func GoRoutineID() (ID string) {
	return getID(string(debug.Stack()))
}

// getID obtains gorutine ID, as of go1.18 a numeric string "1"…
func getID(stackTrace string) (ID string) {
	if !strings.HasPrefix(stackTrace, runtGoroutinePrefix) {
		panic(errors.New("runt.getID: stack trace not starting with: " + runtGoroutinePrefix))
	}
	IDIndex := len(runtGoroutinePrefix)
	spaceIndex := strings.Index(stackTrace[IDIndex:], "\x20")
	if spaceIndex == -1 {
		panic(errors.New("runt.getID: bad stack trace string"))
	}
	return stackTrace[IDIndex : spaceIndex+IDIndex]
}
