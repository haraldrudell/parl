/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package runt

import (
	"runtime/debug"
	"strings"
)

/*
A stack has a leading line with go routine ID, then two lines per frame:
goroutine 1 [running]:
runtime/debug.Stack(0x1, 0x1, 0x2)
        /usr/local/Cellar/go/1.16.6/libexec/src/runtime/debug/stack.go:24 +0x9f
*/
const (
	linesPerStackFrame        = 2
	skipFrames                = 2 // debug.Stack() + Invocation()
	stackRegexpGoroutineIndex = 1
	stackRegexpRestIndex      = 3
	goroutinePrefix           = "goroutine "
	goroutineSuffix           = ":"
	goroutineEndBracket       = "]"
	goroutineStartBracket     = "["
	goRoutineSpaceSuffix      = "\x20"
)

// Invocation returns an invocation stack trace for debug printing, empty string on troubles
// "goroutine 1 [running]:\ngithub.com/haraldrudell/parl/mains.(*Executable).AddErr(0x1809300, 0x158b620, 0xc000183800, 0x1) mains.(*Executable).AddErr-executable.go:302…"
func Invocation(stackFramesToSkip int) (s string) {
	if stackFramesToSkip < 0 {
		stackFramesToSkip = 0
	}
	linesToSkip := (stackFramesToSkip + skipFrames) * linesPerStackFrame
	remainingStackTrace := string(debug.Stack())
	var goRoutineText string
	if index := strings.Index(remainingStackTrace, "\n"); index >= 0 {
		goRoutineText = remainingStackTrace[:index] // "goroutine 1 [running]:"
		remainingStackTrace = remainingStackTrace[index+1:]
	}
	for linesToSkip > 0 {
		if index := strings.Index(remainingStackTrace, "\n"); index >= 0 {
			remainingStackTrace = remainingStackTrace[index+1:]
		}
		linesToSkip--
	}
	return goRoutineText + "\n" + remainingStackTrace
}

func GoRoutineID() (ID string) {
	stackTraceS := string(debug.Stack())
	var firstLine string
	if index := strings.Index(stackTraceS, "\n"); index >= 0 { // first line: "goroutine 1 [running]:"
		firstLine = stackTraceS[:index]
	}
	firstLine = strings.TrimPrefix(firstLine, goroutinePrefix) // leading "goroutine "
	firstLine = strings.TrimSuffix(firstLine, goroutineSuffix) // trailing colon
	if strings.HasSuffix(firstLine, goroutineEndBracket) {     // remove possible ending bracketed text "[…]"
		if index := strings.LastIndex(firstLine, goroutineStartBracket); index >= 0 {
			firstLine = firstLine[:index]
		}
	}
	return strings.TrimSuffix(firstLine, goRoutineSpaceSuffix) // ending space
}
