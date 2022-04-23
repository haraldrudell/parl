/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// debug.Stack uses this prefix in the first line of the result
	runFirstRexpT             = "^goroutine ([[:digit:]]+) [[]([^]]+)[]]:$"
	runtCreatedByPrefix       = "created by "
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
	ID parl.ThreadID
	// Status is typically word “running”
	Status parl.ThreadStatus
	// IsMainThread indicates if this is the thread that launched main.main
	IsMainThread bool
	// Frames is a list of code locations for this thread.
	// [0] is the invoker of goid.NewStack().
	// last is the function starting this thread.
	// Frame.Args is invocation values like "(0x14000113040)".
	Frames []Frame
	// Creator is the code location of the go statement launching
	// this thread.
	// FuncName is "main.main()" for main thread
	Creator pruntime.CodeLocation
}

type Frame struct {
	pruntime.CodeLocation
	// args like "(1, 2, 3)"
	Args string
}

var firstRexp = regexp.MustCompile(runFirstRexpT)

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
	trace := strings.Split(strings.TrimSuffix(string(debug.Stack()), "\n"), "\n")

	// check trace length
	minLines := runtDebugAndStackFrames*runtLinesPerFrame + runtStatusAndCreatorLines
	if len(trace) < minLines || len(trace)&1 == 0 {
		panic(fmt.Errorf("pruntime.Stack trace less than %d lines or even: %d", minLines, len(trace)))
	}

	// first line: s.ID s.Status
	if s.ID, s.Status, err = ParseFirstLine(trace[0]); err != nil {
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

		// get a pointer into the slice where to store
		s.Frames = append(s.Frames, Frame{})
		framep := &s.Frames[len(s.Frames)-1]

		// parse function line
		framep.CodeLocation.FuncName, framep.Args = ParseFuncLine(trace[i])

		// parse file line
		framep.CodeLocation.File, framep.CodeLocation.Line = ParseFileLine(trace[i+1])
	}

	// populate s.IsMainThread s.Creator
	// last 2 lines
	s.Creator.FuncName, s.IsMainThread = ParseCreatedLine(trace[len(trace)-2])
	s.Creator.File, s.Creator.Line = ParseFileLine(trace[len(trace)-1])

	stack = &s
	return
}

// getID obtains gorutine ID, as of go1.18 a numeric string "1"…
func ParseFirstLine(debugStack string) (ID parl.ThreadID, status parl.ThreadStatus, err error) {

	// remove possible lines 2…
	if index := strings.Index(debugStack, "\n"); index != -1 {
		debugStack = debugStack[:index]
	}

	// find ID and status
	matches := firstRexp.FindAllStringSubmatch(debugStack, -1)
	if matches == nil {
		err = perrors.Errorf("goid.ParseFirstStackLine failed to parse: %q", debugStack)
		return
	}

	// return values
	values := matches[0][1:]
	ID = parl.ThreadID(values[0])
	status = parl.ThreadStatus(values[1])

	return
}

// ParseFileLine parses a line of a tab character then absolue file path,
// followed by a colon and line number, then a space character and
// a byte offset
//  "\t/gp-debug-stack/debug-stack.go:29 +0x44"
func ParseFileLine(fileLine string) (file string, line int) {
	var hasTab bool
	var lastColon = -1
	var spaceAfterColon = -1
	if len(fileLine) > 0 {
		hasTab = fileLine[:1] == "\t"
		lastColon = strings.LastIndex(fileLine, ":")
		spaceAfterColon = strings.LastIndex(fileLine, "\x20")
	}
	if !hasTab || lastColon == -1 || spaceAfterColon < lastColon {
		panic(perrors.Errorf("bad debug.Stack: file line: %q", fileLine))
	}

	var err error
	if line, err = strconv.Atoi(fileLine[lastColon+1 : spaceAfterColon]); err != nil {
		panic(perrors.Errorf("bad debug.Stack file line number: %w %q", err, fileLine))
	}
	if line < 1 {
		panic(perrors.Errorf("bad debug.Stack file line number <1: %q", fileLine))
	}

	file = fileLine[1:lastColon]

	return
}

// ParseCreatedLine parses the second-to-last line of the stack trace.
// samples:
//  created by main.main
//  created by main.(*MyType).goroutine1
//  main.main()
func ParseCreatedLine(createdLine string) (funcName string, IsMainThread bool) {

	// if its starts with created, it is a goroutine
	if strings.HasPrefix(createdLine, runtCreatedByPrefix) {
		funcName = createdLine[len(runtCreatedByPrefix):]
		return
	}

	// it is main.main()
	funcName, _ = ParseFuncLine(createdLine)
	IsMainThread = true

	return
}

// ParseFuncLine parses a line of a package name, optionally fully qualified, and
// a possible receiver type name and a function name, followed by a parenthesised
// argument list.
// samples:
//  main.main()
//  main.(*MyType).goroutine1(0x0?, 0x140000120d0, 0x2)
//  codeberg.org/haraldrudell/goprogramming/std/runtime-debug/gp-debug-stack/mypackage.Fn(...)
func ParseFuncLine(funcLine string) (funcName string, args string) {
	leftIndex := strings.Index(funcLine, "(")
	if leftIndex < 1 {
		panic(perrors.Errorf("Bad debug.Stack function line: no left parenthesis: %q", funcLine))
	}

	// determine if parenthesis is for optional type name rarther than function arguments
	if funcLine[leftIndex-1:leftIndex] == "." {
		nextIndex := strings.Index(funcLine[leftIndex+1:], "(")
		if nextIndex < 1 {
			panic(perrors.Errorf("Bad debug.Stack function line: no second left parenthesis: %q", funcLine))
		}
		leftIndex += nextIndex + 1
	}

	funcName = funcLine[:leftIndex]
	args = funcLine[leftIndex:]

	return
}
