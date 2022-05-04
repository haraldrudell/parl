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
	// runtStatusLines is a single line at the beginning of the stack trace
	runtStatusLines = 1
	// runtDebugStackLines are lines for the debug.Stack stack frame
	runtDebugStackLines = 2
	// runtNewStackLines are lines for the goid.NewStack stack frame
	runtNewStackLines = 2
	// runtCreatorLines are 2 optional lines at the end of the stack trace
	runtCreatorLines = 2
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
	// line 1 is status line
	// line 2 is debug.Stack frame
	// created by is 2 optional lines at end
	trace := strings.Split(strings.TrimSuffix(string(debug.Stack()), "\n"), "\n")
	traceLen := len(trace)
	skipAtStart := runtStatusLines + runtDebugStackLines + runtNewStackLines
	skipAtEnd := 0

	// parse possible "created by"
	if traceLen >= runtCreatorLines {
		// populate s.IsMainThread s.Creator
		// last 2 lines
		if s.creator.FuncName, s.isMainThread = ParseCreatedLine(trace[traceLen-2]); !s.isMainThread {
			skipAtEnd += runtCreatorLines
			s.creator.File, s.creator.Line = ParseFileLine(trace[traceLen-1])
		}
	}

	// check trace length: must be at least one frame available
	minLines := skipAtStart + skipAtEnd + // skip lines at beginning and end
		runtLinesPerFrame // one frame available
	if traceLen < minLines || traceLen&1 == 0 {
		panic(fmt.Errorf("pruntime.Stack trace less than %d[%d–%d] lines or even: %d\nTRACE: %s\n",
			minLines, skipAtStart, skipAtEnd, len(trace),
			strings.Join(trace, "\n"),
		))
	}

	// check skipFrames
	maxSkipFrames := (traceLen - minLines) / runtLinesPerFrame
	if skipFrames > maxSkipFrames {
		panic(fmt.Errorf("pruntime.Stack bad skipFrames: %d trace-length: %d[%d–%d] max-skipFrames: %d\nTRACE: %s\n",
			skipFrames, traceLen, skipAtStart, skipAtEnd, maxSkipFrames,
			strings.Join(trace, "\n"),
		))
	}
	skipAtStart += skipFrames * runtLinesPerFrame // remove frames from skipFrames
	skipIndex := traceLen - skipAtEnd             // limit index at end

	// parse first line: s.ID s.Status
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
	for i := skipAtStart; i < skipIndex; i += runtLinesPerFrame {

		var frame Frame

		// parse function line
		frame.CodeLocation.FuncName, frame.args = ParseFuncLine(trace[i])

		// parse file line
		frame.CodeLocation.File, frame.CodeLocation.Line = ParseFileLine(trace[i+1])
		s.frames = append(s.frames, &frame)
	}

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

func (st *Stack) String() (s string) {
	sL := make([]string, 2*len(st.frames))
	for i, frame := range st.frames {
		sL[i] = frame.String()
	}
	if s = strings.Join(sL, "\n"); s != "" {
		s += "\n"
	}
	return fmt.Sprintf("ID: %s IsMain: %t status: %s\n"+
		"%s"+
		"cre: %s",
		st.threadID, st.isMainThread, st.status,
		s,
		st.creator.Long(),
	)
}
