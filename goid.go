/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"regexp"
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	// debug.Stack uses this prefix in the first line of the result
	runFirstRexpT = "^goroutine ([[:digit:]]+) .*$"
)

var firstRexp = regexp.MustCompile(runFirstRexpT)

func goID() (threadID ThreadID) {
	var debugStack = pruntime.StackString()
	if index := strings.Index(debugStack, "\n"); index != -1 {
		debugStack = debugStack[:index-1]
	}
	if matches := firstRexp.FindAllStringSubmatch(debugStack, -1); matches != nil {
		values := matches[0][1:]
		threadID = ThreadID(values[0])
	}
	return
}
