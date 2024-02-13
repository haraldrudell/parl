/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl/pruntime/pruntimelib"
)

// Stackr is a parl-free [pdebug.Stack]
//   - Go stack traces are created by [runtime.Stack] and is a byte slice
//   - [debug.Stack] repeatedly calls [runtime.Stack] with an increased
//     buffer size that is eventually returned
//   - [debug.PrintStack] writes the byte stream to [os.Stderr]
//   - interning large strings is a temporary memory leak.
//     Converting the entire byte-slice stack-trace to string
//     retains the memory for as long as there is a reference to any one character.
//     This leads to megabytes of memory leaks
type StackR struct {
	// ThreadID is a unqique ID associated with this thread.
	// typically numeric string “1”…
	//	- [constraints.Ordered] [fmt.Stringer] [ThreadID.IsValid]
	ThreadID uint64
	// Status is typically word “running”
	Status string
	// isMainThread indicates if this is the thread that launched main.main
	//	- if false, the stack trace is for a goroutine,
	//		a thread directly or indirectly launched by the main thread
	isMainThread bool
	// Frames is a list of code locations for this thread.
	//	- [0] is the most recent code location, typically the invoker of [pdebug.Stack]
	//	- last is the function starting this thread or in the Go runtime
	//	- Frame.Args is invocation values like "(0x14000113040)"
	frames []Frame
	// goFunction is the function used in a go statement
	//	- if isMain is true, it is the zero-value
	goFunction CodeLocation
	// Creator is the code location of the go statement launching
	// this thread
	//	- FuncName is "main.main()" for main thread
	Creator CodeLocation
	// possible ID of creating goroutine
	CreatorID uint64
	// creator goroutine reference “in goroutine 1”
	GoroutineRef string
}

var _ Stack = &StackR{}

// NewStack populates a Stack object with the current thread
// and its stack using debug.Stack
func NewStack(skipFrames int) (stack Stack) {
	var err error
	if skipFrames < 0 {
		skipFrames = 0
	}
	// result of parsing to be returned
	var s StackR

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
	//	- [bytes.TrimSuffix]: no allocations
	//	- [bytes.Split] allocates the [][]byte slice index
	//	- each conversion to string causes an allocation
	var trace = bytes.Split(bytes.TrimSuffix(StackTrace(), []byte{'\n'}), []byte{'\n'})
	var nonEmptyLineCount = len(trace)
	var skipAtStart = runtStatus + runtPdebugStack + runtPruntimeStackTrace
	var skipAtEnd = 0

	// parse possible “created by” line-pair at end
	//	- goroutine creator may be in the two last text lines of the stack trace
	if nonEmptyLineCount-skipAtStart >= runtCreator {
		// the index in trace of created-by line
		var creatorIndex = nonEmptyLineCount - runtCreator
		// temporary creator code location
		var creator CodeLocation
		var goroutineRef string
		// determine s.isMainThread
		creator.FuncName, goroutineRef, s.isMainThread = pruntimelib.ParseCreatedLine(trace[creatorIndex])
		// if a goroutine, store creator
		if !s.isMainThread {
			s.GoroutineRef = goroutineRef
			creator.File, creator.Line = pruntimelib.ParseFileLine(trace[creatorIndex+runtFileLineOffset])
			s.Creator = creator
			// “in goroutine 1”
			if index := strings.LastIndex(goroutineRef, "\x20"); index != -1 {
				var i, _ = strconv.ParseUint(goroutineRef[index+1:], 10, 64)
				if i > 0 {
					s.CreatorID = i
				}
			}
			skipAtEnd += runtCreator
		}

		// if not main thread, store goroutine function
		if !s.isMainThread && creatorIndex >= skipAtStart+runtGoFunction {
			// the trace index for goroutine function
			var goIndex = creatorIndex - runtGoFunction
			s.goFunction.FuncName, _ = pruntimelib.ParseFuncLine(trace[goIndex])
			s.goFunction.File, s.goFunction.Line = pruntimelib.ParseFileLine(trace[goIndex+runtFileLineOffset])
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
	var threadID uint64
	var status string
	if threadID, status, err = pruntimelib.ParseFirstLine(trace[0]); err != nil {
		panic(err)
	}
	s.ThreadID = threadID
	s.Status = status
	//s.SetID(threadID, status)

	var frameCount = (skipIndex - skipAtStart) / runtLinesPerFrame
	if frameCount > 0 {
		// extract the desired stack frames into s.Frames
		// stack:
		//  first line
		//  two lines of runtime/debug.Stack()
		//  two lines of goid.NewStack()
		//  additional frame line-pairs
		//  two lines of goroutine Creator
		var frames = make([]Frame, frameCount)
		var frameStructs = make([]FrameR, frameCount)
		for i, frameIndex := skipAtStart, 0; i < skipIndex; i += runtLinesPerFrame {
			var frame = &frameStructs[frameIndex]

			// parse function line
			frame.CodeLocation.FuncName, frame.args = pruntimelib.ParseFuncLine(trace[i])

			// parse file line
			frame.CodeLocation.File, frame.CodeLocation.Line = pruntimelib.ParseFileLine(trace[i+1])
			frames[frameIndex] = frame
			frameIndex++
		}
		s.frames = frames
	}
	stack = &s

	return
}

// A list of code locations for this thread
//   - index [0] is the most recent code location, typically the invoker requesting the stack trace
//   - includes invocation argument values
func (s *StackR) Frames() (frames []Frame) { return s.frames }

// the goroutine function used to launch this thread
//   - if IsMain is true, zero-value. Check using GoFunction().IsSet()
//   - never nil
func (s *StackR) GoFunction() (goFunction *CodeLocation) { return &s.goFunction }

// true if the thread is the main thread
//   - false for a launched goroutine
func (s *StackR) IsMain() (isMainThread bool) { return s.isMainThread }

// Shorts lists short code locations for all stack frames, most recent first:
// Shorts("prepend") →
//
//	prepend Thread ID: 1
//	prepend main.someFunction()-pruntime.go:84
//	prepend main.main()-pruntime.go:52
func (s *StackR) Shorts(prepend string) (shorts string) {
	if prepend != "" {
		prepend += "\x20"
	}
	sL := []string{
		prepend + "Thread ID: " + strconv.FormatUint(s.ThreadID, 10),
	}
	for _, frame := range s.frames {
		sL = append(sL, prepend+frame.Loc().Short())
	}
	if s.Creator.IsSet() {
		var s3 = prepend + "creator: " + s.Creator.Short()
		if s.GoroutineRef != "" {
			s3 += "\x20" + s.GoroutineRef
		}
		sL = append(sL, s3)
	}
	return strings.Join(sL, "\n")
}

func (s *StackR) Dump() (s2 string) {
	if s == nil {
		return "<nil>"
	}
	var f = make([]string, len(s.frames))
	for i, frame := range s.frames {
		f[i] = fmt.Sprintf("%d: %s Args: %q", i+1, frame.Loc().Dump(), frame.Args())
	}
	return fmt.Sprintf("threadID %d status %q isMain %t frames %d[\n%s\n]\ngoFunction:\n%s\ncreator: ref: %q\n%s",
		s.ThreadID, s.Status, s.IsMain(),
		len(s.frames), strings.Join(f, "\n"),
		s.goFunction.Dump(),
		s.GoroutineRef,
		s.Creator.Dump(),
	)
}

// String is a multi-line stack trace, most recent code location first:
//
//	ID: 18 IsMain: false status: running␤
//	main.someFunction({0x100dd2616, 0x19})␤
//	␠␠pruntime.go:64␤
//	cre: main.main-pruntime.go:53␤
func (s *StackR) String() (s2 string) {

	// convert frames to string slice
	var frames = make([]string, len(s.frames))
	for i, frame := range s.frames {
		frames[i] = frame.String()
	}

	// parent information for goroutine
	var parentS string
	if !s.isMainThread {
		parentS = fmt.Sprintf("\nParent-ID: %d go: %s",
			s.CreatorID,
			s.Creator,
		)
	}

	return fmt.Sprintf("ID: %d status: ‘%s’\n%s%s",
		s.ThreadID, s.Status,
		strings.Join(frames, "\n"),
		parentS,
	)
}

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
