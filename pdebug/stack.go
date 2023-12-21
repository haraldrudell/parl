/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pdebug provides a portable, parsed stack trace.
package pdebug

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// each stack frame is two lines:
	//	- function line
	//	- filename line with leading tab, line number and byte-offset
	runtLinesPerFrame = 2
	// the filename line is 1 line after function line
	runtFileLineOffset = 1
	// runtStatus is a single line at the beginning of the stack trace
	runtStatus = 1
	// runtPdebugStack are lines for the [pdebug.Stack] stack frame
	runtPdebugStack = runtLinesPerFrame
	// runtPruntimeStackTrace are lines for the [pruntime.StackTrace] stack frame
	runtPruntimeStackTrace = runtLinesPerFrame
	// runtCreator are 2 optional lines at the end of the stack trace
	runtCreator = runtLinesPerFrame
	// runtGoFunction are 2 optional lines that precedes runtCreator
	runtGoFunction = runtLinesPerFrame
)

//   - Go stack traces are created by [runtime.Stack] and is a byte slice
//   - [debug.Stack] repeatedly calls [runtime.Stack] with an increased
//     buffer size that is eventually returned
//   - [debug.PrintStack] writes the byte stream to [os.Stderr]
//   - interning large strings is a temporary memory leak.
//     Converting the entire byte-slice stack-trace to string
//     retains the memory for as long as there is a reference to any one character.
//     This leads to megabytes of memory leaks
type Stack struct {
	// ThreadID is a unqique ID associated with this thread.
	// typically numeric string “1”…
	//	- Ordered so it can be used as map key and sorted
	//	- String method converts to string value
	//	- IsValid method determines if not zero-value
	threadID parl.ThreadID
	// Status is typically word “running”
	status parl.ThreadStatus
	// isMainThread indicates if this is the thread that launched main.main
	//	- if false, the stack trace is for a goroutine
	isMainThread bool
	// Frames is a list of code locations for this thread.
	//	- [0] is the most recent code location, typically the invoker of [pdebug.Stack]
	//	- last is the function starting this thread or in the Go runtime
	//	- Frame.Args is invocation values like "(0x14000113040)"
	frames []parl.Frame
	// goFunction is the function used in a go statement
	//	- if isMain is true, it is the zero-value
	goFunction pruntime.CodeLocation
	// Creator is the code location of the go statement launching
	// this thread
	//	- FuncName is "main.main()" for main thread
	creator pruntime.CodeLocation
	// creator goroutine reference
	goroutineRef string
}

// NewStack populates a Stack object with the current thread
// and its stack using debug.Stack
func NewStack(skipFrames int) (stack parl.Stack) {
	var err error
	if skipFrames < 0 {
		skipFrames = 0
	}
	// result of parsing to be returned
	var s Stack

	// [pruntime.StackTrace] returns a stack trace with final newline:
	// goroutine␠18␠[running]:
	// github.com/haraldrudell/parl/pruntime.StackTrace()
	// ␉/opt/sw/parl/pruntime/stack-trace.go:24␠+0x50
	// github.com/haraldrudell/parl/pruntime.TestStackTrace(0x14000122820)
	// ␉/opt/sw/parl/pruntime/stack-trace_test.go:14␠+0x20
	// testing.tRunner(0x14000122820,␠0x104c204c8)
	// ␉/opt/homebrew/Cellar/go/1.21.4/libexec/src/testing/testing.go:1595␠+0xe8
	// created␠by␠testing.(*T).Run␠in␠goroutine␠1
	// ␉/opt/homebrew/Cellar/go/1.21.4/libexec/src/testing/testing.go:1648␠+0x33c

	// trace is array of byte-slice lines: removed final newline and split on newline
	//	- trace[0] is status line “goroutine␠18␠[running]:”
	//	- trace[1…2] is [pruntime.StackTrace] frame
	//	- trace[3…4] is [pdebug.Stack] frame
	//	- final newline is removed, so last line is non-empty
	//	- created by is 2 optional lines at end
	//	- if created by is present, the two preceding lines is the goroutine function
	var trace = bytes.Split(bytes.TrimSuffix(pruntime.StackTrace(), []byte{'\n'}), []byte{'\n'})
	var nonEmptyLineCount = len(trace)
	var skipAtStart = runtStatus + runtPdebugStack + runtPruntimeStackTrace
	var skipAtEnd = 0

	// parse possible “created by” line-pair at end
	//	- goroutine creator may be in the two last text lines of the stack trace
	if nonEmptyLineCount-skipAtStart >= runtCreator {
		// the index in trace of created-by line
		var creatorIndex = nonEmptyLineCount - runtCreator
		// temporary creator code location
		var creator pruntime.CodeLocation
		var goroutineRef string
		// determine s.isMainThread
		creator.FuncName, goroutineRef, s.isMainThread = ParseCreatedLine(trace[creatorIndex])
		// if a goroutine, store creator
		if !s.isMainThread {
			s.goroutineRef = goroutineRef
			creator.File, creator.Line = ParseFileLine(trace[creatorIndex+runtFileLineOffset])
			s.creator = creator
			skipAtEnd += runtCreator
		}

		// if not main thread, store goroutine function
		if !s.isMainThread && creatorIndex >= skipAtStart+runtGoFunction {
			// the trace index for goroutine function
			var goIndex = creatorIndex - runtGoFunction
			s.goFunction.FuncName, _ = ParseFuncLine(trace[goIndex])
			s.goFunction.File, s.goFunction.Line = ParseFileLine(trace[goIndex+runtFileLineOffset])
		}
	}

	// check trace length: must be at least one frame available
	var minLines = skipAtStart + skipAtEnd + // skip lines at beginning and end
		runtLinesPerFrame // one frame available
	if nonEmptyLineCount < minLines || nonEmptyLineCount&1 == 0 {
		panic(fmt.Errorf("pdebug.Stack trace less than %d[%d–%d] lines or even: %d\nTRACE: %s%s",
			minLines, skipAtStart, skipAtEnd, len(trace),
			string(bytes.Join(trace, []byte{'\n'})),
			"\n",
		))
	}

	// check skipFrames
	var maxSkipFrames = (nonEmptyLineCount - minLines) / runtLinesPerFrame
	if skipFrames > maxSkipFrames {
		panic(fmt.Errorf("pruntime.Stack bad skipFrames: %d trace-length: %d[%d–%d] max-skipFrames: %d\nTRACE: %s%s",
			skipFrames, nonEmptyLineCount, skipAtStart, skipAtEnd, maxSkipFrames,
			string(bytes.Join(trace, []byte{'\n'})),
			"\n",
		))
	}
	skipAtStart += skipFrames * runtLinesPerFrame // remove frames from skipFrames
	var skipIndex = nonEmptyLineCount - skipAtEnd // limit index at end

	// parse first line: s.ID s.Status
	var threadID parl.ThreadID
	var status parl.ThreadStatus
	if threadID, status, err = ParseFirstLine(trace[0]); err != nil {
		panic(err)
	}
	s.threadID = threadID
	s.status = status
	//s.SetID(threadID, status)

	// extract the desired stack frames into s.Frames
	// stack:
	//  first line
	//  two lines of runtime/debug.Stack()
	//  two lines of goid.NewStack()
	//  additional frame line-pairs
	//  two lines of goroutine Creator
	var frames []parl.Frame
	for i := skipAtStart; i < skipIndex; i += runtLinesPerFrame {

		var frame Frame

		// parse function line
		frame.CodeLocation.FuncName, frame.args = ParseFuncLine(trace[i])

		// parse file line
		frame.CodeLocation.File, frame.CodeLocation.Line = ParseFileLine(trace[i+1])
		frames = append(frames, &frame)
	}
	s.frames = frames
	//s.SetFrames(frames)

	stack = &s
	return
}

var _ parl.Stack = &Stack{}

// thread ID 1… for the thread requesting the stack trace
//   - ThreadID is comparable and has IsValid and String methods
//   - ThreadID is typically an incremented 64-bit integer with
//     main thread having ID 1
func (s *Stack) ID() (threadID parl.ThreadID) { return s.threadID }

// true if the thread is the main thread
//   - false for a launched goroutine
func (s *Stack) IsMain() (isMainThread bool) { return s.isMainThread }

// a word indicating thread status, typically word “running”
func (s *Stack) Status() (status parl.ThreadStatus) { return s.status }

// the goroutine function used to launch this thread
//   - if IsMain is true, zero-value. Check using GoFunction().IsSet()
//   - never nil
func (s *Stack) GoFunction() (goFunction *pruntime.CodeLocation) { return &s.goFunction }

// the code location of the go statement creating this thread
//   - if IsMain is true, zero-value. Check with Creator().IsSet()
//   - never nil
func (s *Stack) Creator() (creator *pruntime.CodeLocation, goRoutineRef string) {
	var c = s.creator
	creator = &c
	goRoutineRef = s.goroutineRef
	return
}

// A list of code locations for this thread
//   - index [0] is the most recent code location, typically the invoker requesting the stack trace
//   - includes invocation argument values
func (s *Stack) Frames() (frames []parl.Frame) { return s.frames }

// func (s *Stack) MostRecentFrame() (frame parl.Frame) {
// 	var f Frame
// 	if len(s.frames) > 0 {
// 		fp := s.frames[0].(*Frame)
// 		f = *fp
// 	}
// 	frame = &f
// 	return
// }

// Shorts lists short code locations for all stack frames, most recent first:
// Shorts("prepend") →
//
//	prepend Thread ID: 1
//	prepend main.someFunction()-pruntime.go:84
//	prepend main.main()-pruntime.go:52
func (s *Stack) Shorts(prepend string) (shorts string) {
	if prepend != "" {
		prepend += "\x20"
	}
	sL := []string{
		prepend + "Thread ID: " + s.threadID.String(),
	}
	for _, frame := range s.frames {
		sL = append(sL, prepend+frame.Loc().Short())
	}
	if s.creator.IsSet() {
		var s3 = prepend + "creator: " + s.creator.Short()
		if s.goroutineRef != "" {
			s3 += "\x20" + s.goroutineRef
		}
		sL = append(sL, s3)
	}
	return strings.Join(sL, "\n")
}

// SetID updates goroutine ID and thread status from the stack trace
// status line
// func (s *Stack) SetID(threadID parl.ThreadID, status parl.ThreadStatus) {
// 	s.threadID = threadID
// 	s.status = status
// }

// SetCreator is used if Stack is for a goroutine
//   - describing whether it is the main thread
//   - if a gouroutine, where the launching go statement was located
// func (s *Stack) SetCreator(creator *pruntime.CodeLocation, isMainThread bool) {
// 	s.creator = *creator
// 	s.isMainThread = isMainThread
// }

//func (s *Stack) SetFrames(frames []parl.Frame) { s.frames = frames }

func (s *Stack) Dump() (s2 string) {
	if s == nil {
		return "<nil>"
	}
	var f = make([]string, len(s.frames))
	for i, frame := range s.frames {
		f[i] = fmt.Sprintf("%d: %s Args: %q", i+1, frame.Loc().Dump(), frame.Args())
	}
	return parl.Sprintf("threadID %q status %q isMain %t frames %d[\n%s\n]\ngoFunction:\n%s\ncreator: ref: %q\n%s",
		s.threadID, s.status, s.IsMain(),
		len(s.frames), strings.Join(f, "\n"),
		s.goFunction.Dump(),
		s.goroutineRef,
		s.creator.Dump(),
	)
}

// String is a multi-line stack trace, most recent code location first:
//
//	ID: 18 IsMain: false status: running␤
//	main.someFunction({0x100dd2616, 0x19})␤
//	␠␠pruntime.go:64␤
//	cre: main.main-pruntime.go:53␤
func (s *Stack) String() (s2 string) {
	sL := make([]string, len(s.frames))
	for i, frame := range s.frames {
		sL[i] = frame.String()
	}
	if s2 = strings.Join(sL, "\n"); s2 != "" {
		s2 += "\n"
	}
	var cre = s.creator.Long()
	if s.goroutineRef != "" {
		cre += "\x20" + s.goroutineRef
	}
	return fmt.Sprintf("ID: %s IsMain: %t status: %s\n"+
		"%s"+
		"cre: %s",
		s.threadID, s.isMainThread, s.status,
		s2,
		cre,
	)
}
