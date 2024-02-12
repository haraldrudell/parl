/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime2

import (
	"bytes"
	"fmt"
	"strconv"
)

const (
	fileLineTabRune = '\t'
)

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
		panic(fmt.Errorf("bad debug.Stack: file line: %q", string(fileLine)))
	}

	var err error
	if line, err = strconv.Atoi(string(fileLine[lastColon+1 : spaceAfterColon])); err != nil {
		panic(fmt.Errorf("bad debug.Stack file line number: %w %q", err, string(fileLine)))
	} else if line < 1 {
		panic(fmt.Errorf("bad debug.Stack file line number <1: %q", string(fileLine)))
	}

	// absolute filename
	file = string(fileLine[1:lastColon])

	return
}
