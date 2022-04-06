/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	// debug.Stack uses this prefix in the first line of the result
	runtGoroutinePrefix       = "goroutine "
	runtStackStatusLeft       = "["
	runtStackStatusRight      = "]"
	runtCreatedByPrefix       = "created by "
	runtDebugAndStackFrames   = 2
	runtStatusAndCreatorLines = 3
	runtLinesPerFrame         = 2
	runtStatusLineLength      = 1
	runtCreatorLines          = 2
)

type Stack struct {
	// ThreadID is unqique for this thread.
	// typically numeric string “1”…
	ThreadID string
	// Status is typically word “running”
	Status string
	// IsMainThread indicates if this is the thread that laucnhed main.main
	IsMainThread bool
	// Frame.Args like "(0x14000113040)".
	Frames []Frame
	// Creator.FuncName is "main.main()" for main thread
	Creator    CodeLocation
	DebugStack CodeLocation
}

type Frame struct {
	CodeLocation
	// args like "(1, 2, 3)"
	Args string
}

// NewStack populates a Stack object with the current thread
// and its stack
func NewStack(skipFrames int) (stack *Stack) {
	var err error
	if skipFrames < 0 {
		skipFrames = 0
	}
	var s Stack
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

	// first line
	if s.ThreadID, s.Status, err = ParseFirstStackLine(trace[0], false); err != nil {
		panic(err)
	}

	// stack
	firstIndex := runtStatusLineLength + (skipFrames+runtDebugAndStackFrames)*runtLinesPerFrame
	for i := firstIndex; i < len(trace)-runtCreatorLines; i += runtLinesPerFrame {
		var frame *Frame
		if frame, err = ParseStackFrame(trace[i:i+2], false); err != nil {
			panic(err)
		}
		s.Frames = append(s.Frames, *frame)
	}

	// last 2 lines
	twoLines := trace[len(trace)-2:]
	if strings.HasPrefix(twoLines[0], runtCreatedByPrefix) {
		twoLines[0] = twoLines[0][len(runtCreatedByPrefix):]
	} else {
		s.IsMainThread = true
	}
	var frame *Frame
	if frame, err = ParseStackFrame(twoLines, true); err != nil {
		panic(err)
	}
	s.Creator = frame.CodeLocation

	stack = &s
	return
}

// getID obtains gorutine ID, as of go1.18 a numeric string "1"…
func ParseFirstStackLine(stackTrace string, onlyID bool) (ID string, status string, err error) {

	// get ID
	if !strings.HasPrefix(stackTrace, runtGoroutinePrefix) {
		err = errors.New("runt.getID: stack trace not starting with: " + runtGoroutinePrefix)
		return
	}
	IDIndex := len(runtGoroutinePrefix)
	spaceIndex := strings.Index(stackTrace[IDIndex:], "\x20")
	if spaceIndex == -1 {
		err = errors.New("runt.getID: bad stack trace string")
		return
	}
	ID = stackTrace[IDIndex : spaceIndex+IDIndex]
	if onlyID {
		return
	}

	// get status
	line0 := strings.Split(stackTrace, "\n")[0]
	line := line0
	if len(line) >= spaceIndex {
		line = line[spaceIndex:]
	}
	left := strings.Index(line, runtStackStatusLeft)
	right := strings.Index(line, runtStackStatusRight)
	if left == -1 || right == -1 || left >= right {
		err = fmt.Errorf("runt.getID: unparseable first line: %q", line0)
		return
	}
	status = line[left:right]
	return
}

func ParseStackFrame(twoLines []string, noArgs bool) (frame *Frame, err error) {
	f := Frame{}
	length := len(twoLines)
	if length != 2 {
		err = errors.Errorf("pruntime.Stack: input length not 2: %d", length)
		return
	}
	fn := twoLines[0]
	file := twoLines[1]
	// find “last” left parenthesis
	left := strings.Index(fn, "(")
	restLeft := 0
	if left > 0 && fn[left-1:left] == "." { // it’s a type before the function name
		rest := fn[left+1:]
		if restLeft = strings.Index(rest, "("); restLeft != -1 {
			left += restLeft + 1
		}
	}
	if restLeft == -1 && noArgs {
		left = len(fn)
		restLeft = 0
	}
	if restLeft == -1 {
		err = errors.Errorf("pruntime.Stack: bad function line: %q", fn)
		return
	}
	f.CodeLocation.FuncName = fn[:left]
	f.Args = fn[left:]
	var hasTab bool
	var lastColon = -1
	var spaceAfterColon = -1
	if len(file) > 0 {
		hasTab = file[0:1] == "\t"
		lastColon = strings.LastIndex(file, ":")
		spaceAfterColon = strings.Index(file[lastColon+1:], "\x20")
	}
	if !hasTab || lastColon == -1 || spaceAfterColon == -1 {
		err = fmt.Errorf("pruntime.Stack: bad file line: %q", file)
		return
	}
	f.CodeLocation.File = file[1:lastColon]
	f.CodeLocation.Line, err = strconv.Atoi(file[lastColon+1 : lastColon+1+spaceAfterColon])
	if f.CodeLocation.Line < 1 && err == nil {
		err = fmt.Errorf("pruntime.Stack: bad file line number: %q", file)
		return
	} else if err != nil {
		err = fmt.Errorf("pruntime.Stack: file line parse failed: %q", file)
		return
	}
	frame = &f
	return
}
