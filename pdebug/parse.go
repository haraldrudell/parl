/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pdebug

import (
	"bytes"
	"strconv"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	fileLineTabRune = '\t'
)

// created by begins lines returned by runtime.Stack if
// the executing thread is a launched goroutine “created by ”
var runtCreatedByPrefix = []byte("created by ")

// getID obtains gorutine ID, as of go1.18 a numeric string "1"…
func ParseFirstLine(debugStack []byte) (ID parl.ThreadID, status parl.ThreadStatus, err error) {

	var uID uint64
	var status0 string
	if uID, status0, err = pruntime.ParseFirstLine(debugStack); err != nil {
		err = perrors.Stack(err)
		return
	}
	ID = parl.ThreadID(uID)
	status = parl.ThreadStatus(status0)

	return
}

// ParseFileLine parses a line of a tab character then absolue file path,
// followed by a colon and line number, then a space character and
// a byte offset.
//
//	"\t/gp-debug-stack/debug-stack.go:29 +0x44"
//	"\t/opt/sw/parl/g0/waiterr.go:49"
func ParseFileLine(fileLine []byte) (file string, line int) {
	var hasTab bool
	var lastColon = -1
	var spaceAfterColon = -1
	if len(fileLine) > 0 {
		hasTab = fileLine[0] == fileLineTabRune
		lastColon = bytes.LastIndexByte(fileLine, ':')
		if spaceAfterColon = bytes.LastIndexByte(fileLine, '\x20'); spaceAfterColon == -1 {
			spaceAfterColon = len(fileLine)
		}
	}
	if !hasTab || lastColon == -1 || spaceAfterColon < lastColon {
		panic(perrors.Errorf("bad debug.Stack: file line: %q", string(fileLine)))
	}

	var err error
	if line, err = strconv.Atoi(string(fileLine[lastColon+1 : spaceAfterColon])); err != nil {
		panic(perrors.Errorf("bad debug.Stack file line number: %w %q", err, string(fileLine)))
	}
	if line < 1 {
		panic(perrors.Errorf("bad debug.Stack file line number <1: %q", string(fileLine)))
	}

	// absolute filename
	file = string(fileLine[1:lastColon])

	return
}

// ParseCreatedLine parses the second-to-last line of the stack trace.
// samples:
//   - “created by main.main”
//   - “created by main.(*MyType).goroutine1”
//   - “main.main()”
//   - go1.21.5 231219: “created by codeberg.org/haraldrudell/tools/gact.(*Transcriber).TranscriberThread in goroutine 9”
func ParseCreatedLine(createdLine []byte) (funcName, goroutineRef string, IsMainThread bool) {

	// remove prefix “created by ”
	var remain = bytes.TrimPrefix(createdLine, runtCreatedByPrefix)

	// if the line did not have this prefix, it is the main thread that launched main.main
	if IsMainThread = len(remain) == len(createdLine); IsMainThread {
		return // main thread: IsMainThread: true funcName zero-value
	}

	// “codeberg.org/haraldrudell/tools/gact.(*Transcriber).TranscriberThread in goroutine 9”
	if index := bytes.IndexByte(remain, '\x20'); index != -1 {
		funcName = string(remain[:index])
		goroutineRef = string(remain[index+1:])
	} else {
		funcName = string(remain)
	}

	return
}

// ParseFuncLine parses a line of a package name, optionally fully qualified, and
// a possible receiver type name and a function name, followed by a parenthesised
// argument list.
// samples:
//
//	main.main()
//	main.(*MyType).goroutine1(0x0?, 0x140000120d0, 0x2)
//	codeberg.org/haraldrudell/goprogramming/std/runtime-debug/gp-debug-stack/mypackage.Fn(...)
func ParseFuncLine(funcLine []byte) (funcName string, args string) {
	var leftIndex = bytes.IndexByte(funcLine, '(')
	if leftIndex < 1 {
		panic(perrors.Errorf("Bad debug.Stack function line: no left parenthesis: %q", funcLine))
	}

	// determine if parenthesis is for optional type name rarther than function arguments
	if funcLine[leftIndex-1] == '.' {
		nextIndex := bytes.IndexByte(funcLine[leftIndex+1:], '(')
		if nextIndex < 1 {
			panic(perrors.Errorf("Bad debug.Stack function line: no second left parenthesis: %q", funcLine))
		}
		leftIndex += nextIndex + 1
	}

	funcName = string(funcLine[:leftIndex])
	args = string(funcLine[leftIndex:])

	return
}
