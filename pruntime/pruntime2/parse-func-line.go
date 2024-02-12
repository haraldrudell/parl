/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime2

import (
	"bytes"
	"fmt"
)

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
		panic(fmt.Errorf("bad debug.Stack function line: no left parenthesis: %q", funcLine))
	}

	// determine if parenthesis is for optional type name rarther than function arguments
	if funcLine[leftIndex-1] == '.' {
		nextIndex := bytes.IndexByte(funcLine[leftIndex+1:], '(')
		if nextIndex < 1 {
			panic(fmt.Errorf("bad debug.Stack function line: no second left parenthesis: %q", funcLine))
		}
		leftIndex += nextIndex + 1
	}

	funcName = string(funcLine[:leftIndex])
	args = string(funcLine[leftIndex:])

	return
}
