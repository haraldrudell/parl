/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntimelib

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

const (
	// debug.Stack uses this prefix in the first line of the result
	runFirstRexpT = "^goroutine ([[:digit:]]+) [[]([^]]+)[]]:$"
)

var firstRexp = regexp.MustCompile(runFirstRexpT)

// getID obtains gorutine ID, as of go1.18 a numeric string "1"…
func ParseFirstLine(debugStack []byte) (ID uint64, status string, err error) {

	// remove possible lines 2…
	if index := bytes.IndexByte(debugStack, '\n'); index != -1 {
		debugStack = debugStack[:index]
	}

	// find ID and status
	var matches = firstRexp.FindAllSubmatch(debugStack, -1)
	if matches == nil {
		err = fmt.Errorf("goid.ParseFirstStackLine failed to parse: %q", string(debugStack))
		return
	}

	// return values
	var values = matches[0][1:]
	if ID, err = strconv.ParseUint(string(values[0]), 10, 64); err != nil {
		return
	}
	status = string(values[1])

	return
}
