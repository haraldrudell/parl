/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"runtime/debug"
	"strings"
)

const (
	prLeadInLines        = 1
	prDebugStackStdFrame = 1
	prDebugStackFnFrame  = 1
	prLinesPerFrame      = 2
	prCreatorLines       = 2
)

/*
DebugStack produces a string stack frame modified debug.Stack:
Stack frames other than the callers of puntime.DebugStack are removed.
skipFrames allows for removing additional frames.
tabs are replaces with two spaces.
*/
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
	// convert to string, remove final newline, split into lines
	trace := strings.Split(strings.TrimSuffix(string(debug.Stack()), "\n"), "\n")

	// check skipFrames maximum value
	lineCount := len(trace)
	frameCount := (lineCount-prLeadInLines-prCreatorLines)/prLinesPerFrame -
		prDebugStackStdFrame - prDebugStackFnFrame
	if skipFrames > frameCount {
		skipFrames = frameCount
	}

	// remove undesirable stack frames
	skipLines := prLinesPerFrame * (prDebugStackStdFrame + prDebugStackFnFrame + skipFrames)
	copy(trace[prLeadInLines:], trace[prLeadInLines+skipLines:])
	trace = trace[:len(trace)-skipLines]

	return strings.ReplaceAll(strings.Join(trace, "\n"), "\t", "\x20\x20")
}
