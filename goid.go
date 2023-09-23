/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"bytes"
	"regexp"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// debug.Stack uses this prefix in the first line of the result
	runFirstRexpT = "^goroutine ([[:digit:]]+) .*$"
	newlineByte   = byte('\n')
)

var firstRexp = regexp.MustCompile(runFirstRexpT)

func goID() (threadID ThreadID) {

	// stack trace byte-slice part of large allocation
	var debugStack = pruntime.StackTrace()

	// slice to first line
	if index := bytes.IndexByte(debugStack, newlineByte); index != -1 {
		debugStack = debugStack[:index-1]
	}

	var matches = firstRexp.FindAllSubmatch(debugStack, -1)
	if matches == nil {
		panic(perrors.ErrorfPF("stack trace parse failed: %q", string(debugStack)))
	}
	if u64, err := strconv.ParseUint(string(matches[0][1:][0]), 10, 64); err != nil {
		panic(perrors.ErrorfPF("parse goid failed: %q %w", string(debugStack), err))
	} else {
		threadID = ThreadID(u64)
	}
	return
}
