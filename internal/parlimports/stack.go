/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlimports

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

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
	threadID uint64
	// Status is typically word “running”
	status string
	// isMainThread indicates if this is the thread that launched main.main
	isMainThread bool
	// Frames is a list of code locations for this thread.
	// [0] is the invoker of goid.NewStack().
	// last is the function starting this thread.
	// Frame.Args is invocation values like "(0x14000113040)".
	frames []*Frame
	// Creator is the code location of the go statement launching
	// this thread.
	// FuncName is "main.main()" for main thread
	creator pruntime.CodeLocation
}

// NewStack populates a Stack object with the current thread
// and its stack using debug.Stack
func NewStack(skipFrames int) (stack *Stack) {
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
	stackBytes := debug.Stack()
	stackString := string(stackBytes)
	trace := strings.Split(strings.TrimSuffix(stackString, "\n"), "\n")
	traceLen := len(trace)
	skipAtStart := runtStatusLines + runtDebugStackLines + runtNewStackLines
	skipAtEnd := 0

	// parse possible "created by"
	if traceLen >= runtCreatorLines {
		// populate s.IsMainThread s.Creator
		// last 2 lines
		var isMainThread bool
		var creator pruntime.CodeLocation
		if creator.FuncName, isMainThread = ParseCreatedLine(trace[traceLen-2]); !isMainThread {
			skipAtEnd += runtCreatorLines
			creator.File, creator.Line = ParseFileLine(trace[traceLen-1])
		}
		s.SetCreator(&creator, isMainThread)
	}

	// check trace length: must be at least one frame available
	minLines := skipAtStart + skipAtEnd + // skip lines at beginning and end
		runtLinesPerFrame // one frame available
	if traceLen < minLines || traceLen&1 == 0 {
		panic(fmt.Errorf("pruntime.Stack trace less than %d[%d–%d] lines or even: %d\nTRACE: %s%s",
			minLines, skipAtStart, skipAtEnd, len(trace),
			strings.Join(trace, "\n"),
			"\n",
		))
	}

	// check skipFrames
	maxSkipFrames := (traceLen - minLines) / runtLinesPerFrame
	if skipFrames > maxSkipFrames {
		panic(fmt.Errorf("pruntime.Stack bad skipFrames: %d trace-length: %d[%d–%d] max-skipFrames: %d\nTRACE: %s%s",
			skipFrames, traceLen, skipAtStart, skipAtEnd, maxSkipFrames,
			strings.Join(trace, "\n"),
			"\n",
		))
	}
	skipAtStart += skipFrames * runtLinesPerFrame // remove frames from skipFrames
	skipIndex := traceLen - skipAtEnd             // limit index at end

	// parse first line: s.ID s.Status
	var threadID uint64
	var status string
	if threadID, status, err = ParseFirstLine(trace[0]); err != nil {
		panic(err)
	}
	s.SetID(threadID, status)

	// extract the desired stack frames into s.Frames
	// stack:
	//  first line
	//  two lines of runtime/debug.Stack()
	//  two lines of goid.NewStack()
	//  additional frame line-pairs
	//  two lines of goroutine Creator
	var frames []*Frame
	for i := skipAtStart; i < skipIndex; i += runtLinesPerFrame {

		var frame Frame

		// parse function line
		frame.CodeLocation.FuncName, frame.args = ParseFuncLine(trace[i])

		// parse file line
		frame.CodeLocation.File, frame.CodeLocation.Line = ParseFileLine(trace[i+1])
		frames = append(frames, &frame)
	}
	s.SetFrames(frames)

	stack = &s
	return
}

func (s *Stack) ID() (threadID uint64) {
	return s.threadID
}

func (s *Stack) IsMain() (isMainThread bool) {
	return s.isMainThread
}

func (s *Stack) Status() (status string) {
	return s.status
}

func (s *Stack) Creator() (creator *pruntime.CodeLocation) {
	return &s.creator
}

func (s *Stack) Frames() (frames []*Frame) {
	return s.frames
}

func (st *Stack) Shorts(prepend string) (s string) {
	sL := []string{
		prepend + "Thread ID: " + strconv.FormatUint(st.threadID, 10),
	}
	for _, frame := range st.frames {
		sL = append(sL, prepend+frame.Loc().Short())
	}
	if st.creator.IsSet() {
		sL = append(sL, prepend+"creator: "+st.creator.Short())
	}
	return strings.Join(sL, "\n")
}

func (st *Stack) SetID(threadID uint64, status string) {
	st.threadID = threadID
	st.status = status
}

func (st *Stack) SetCreator(creator *pruntime.CodeLocation, isMainThread bool) {
	st.creator = *creator
	st.isMainThread = isMainThread
}

func (st *Stack) SetFrames(frames []*Frame) {
	st.frames = frames
}

func (st *Stack) String() (s string) {
	sL := make([]string, len(st.frames))
	for i, frame := range st.frames {
		sL[i] = frame.String()
	}
	if s = strings.Join(sL, "\n"); s != "" {
		s += "\n"
	}
	return fmt.Sprintf("ID: %d IsMain: %t status: %s\n"+
		"%s"+
		"cre: %s",
		st.threadID, st.isMainThread, st.status,
		s,
		st.creator.Long(),
	)
}
