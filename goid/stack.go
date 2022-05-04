/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	runtDebugAndStackFrames   = 2
	runtStatusAndCreatorLines = 3
	runtLinesPerFrame         = 2
	runtStatusLineLength      = 1
	runtCreatorLines          = 2
)

type Stack struct {
	// ThreadID is a unqique ID associated with this thread.
	// typically numeric string “1”…
	// it can be used as a map key or converted to string
	threadID parl.ThreadID
	// Status is typically word “running”
	status parl.ThreadStatus
	// isMainThread indicates if this is the thread that launched main.main
	isMainThread bool
	// Frames is a list of code locations for this thread.
	// [0] is the invoker of goid.NewStack().
	// last is the function starting this thread.
	// Frame.Args is invocation values like "(0x14000113040)".
	frames []parl.Frame
	// Creator is the code location of the go statement launching
	// this thread.
	// FuncName is "main.main()" for main thread
	creator pruntime.CodeLocation
}

// NewStack populates a Stack object with the current thread
// and its stack using debug.Stack
func NewStack(skipFrames int) (stack parl.Stack) {
	var err error
	if skipFrames < 0 {
		skipFrames = 0
	}
	var s Stack

	// trace is a stack trace as a string with final newline removed split into lines
	/*
		goroutine 18 [running]:
		runtime/debug.Stack()
			/opt/homebrew/Cellar/go/1.18/libexec/src/runtime/debug/stack.go:24 +0x68
		github.com/haraldrudell/parl/pruntime.NewStack()
		…
		created by testing.(*T).Run
			/opt/homebrew/Cellar/go/1.18/libexec/src/testing/testing.go:1486 +0x300
	*/
	trace := strings.Split(strings.TrimSuffix(string(debug.Stack()), "\n"), "\n")

	// check trace length
	minLines := runtDebugAndStackFrames*runtLinesPerFrame + runtStatusAndCreatorLines
	if len(trace) < minLines || len(trace)&1 == 0 {
		panic(fmt.Errorf("pruntime.Stack trace less than %d lines or even: %d", minLines, len(trace)))
	}

	// first line: s.ID s.Status
	if s.threadID, s.status, err = ParseFirstLine(trace[0]); err != nil {
		panic(err)
	}

	// extract the desired stack frames into s.Frames
	// stack:
	//  first line
	//  two lines of runtime/debug.Stack()
	//  two lines of goid.NewStack()
	//  additional frame line-pairs
	//  two lines of goroutine Creator
	firstIndex := runtStatusLineLength + (skipFrames+runtDebugAndStackFrames)*runtLinesPerFrame
	for i := firstIndex; i < len(trace)-runtCreatorLines; i += runtLinesPerFrame {

		var frame Frame

		// parse function line
		frame.CodeLocation.FuncName, frame.args = ParseFuncLine(trace[i])

		// parse file line
		frame.CodeLocation.File, frame.CodeLocation.Line = ParseFileLine(trace[i+1])
		s.frames = append(s.frames, &frame)
	}

	// populate s.IsMainThread s.Creator
	// last 2 lines
	s.creator.FuncName, s.isMainThread = ParseCreatedLine(trace[len(trace)-2])
	s.creator.File, s.creator.Line = ParseFileLine(trace[len(trace)-1])

	stack = &s
	return
}

func (s *Stack) ID() (threadID parl.ThreadID) {
	return s.threadID
}

func (s *Stack) IsMain() (isMainThread bool) {
	return s.isMainThread
}

func (s *Stack) Status() (status parl.ThreadStatus) {
	return s.status
}

func (s *Stack) Creator() (creator *pruntime.CodeLocation) {
	return &s.creator
}

func (s *Stack) Frames() (frames []parl.Frame) {
	return s.frames
}
