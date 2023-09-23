/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pdebug

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// debug.Stack uses this prefix in the first line of the result
	runFirstRexpT       = "^goroutine ([[:digit:]]+) [[]([^]]+)[]]:$"
	runtCreatedByPrefix = "created by "
	fileLineTabRune     = '\t'
)

var firstRexp = regexp.MustCompile(runFirstRexpT)

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
	var u64 uint64
	if u64, err = strconv.ParseUint(values[0], 10, 64); err != nil {
		return
	}
	ID = parl.ThreadID(u64)
	status = parl.ThreadStatus(values[1])

	return
}

// ParseFileLine parses a line of a tab character then absolue file path,
// followed by a colon and line number, then a space character and
// a byte offset.
//
//	"\t/gp-debug-stack/debug-stack.go:29 +0x44"
//	"\t/opt/sw/parl/g0/waiterr.go:49"
func ParseFileLine(fileLine string) (file string, line int) {
	var hasTab bool
	var lastColon = -1
	var spaceAfterColon = -1
	if len(fileLine) > 0 {
		hasTab = fileLine[0] == fileLineTabRune
		lastColon = strings.LastIndex(fileLine, ":")
		if spaceAfterColon = strings.LastIndex(fileLine, "\x20"); spaceAfterColon == -1 {
			spaceAfterColon = len(fileLine)
		}
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
//
//	created by main.main
//	created by main.(*MyType).goroutine1
//	main.main()
func ParseCreatedLine(createdLine string) (funcName string, IsMainThread bool) {

	// check if created frame exists
	if !strings.HasPrefix(createdLine, runtCreatedByPrefix) {
		IsMainThread = true
		return
	}

	funcName = createdLine[len(runtCreatedByPrefix):]

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
