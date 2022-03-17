/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"fmt"
	"io"
	"runtime"
)

const (
	maxStackFrameSize   = 32
	callersFramesToSkip = 1
)

// StackSlice represents a StackSlice of program counters.
type StackSlice []CodeLocation

// NewStackSlice gets a slice of stack frames
func NewStackSlice(skip int) (slice StackSlice) {

	// get the slice of runtime.Frames
	pcs := make([]uintptr, maxStackFrameSize)
	entries := runtime.Callers(callersFramesToSkip+skip, pcs)
	if entries == 0 {
		return
	}
	frames := runtime.CallersFrames(pcs[:entries])

	// convert to slice of CodeLocation
	for {
		frame, more := frames.Next()
		slice = append(slice, *GetCodeLocation(&frame))
		if !more {
			break
		}
	}

	return
}

func (st StackSlice) Short() (s string) {
	s = " at " + st[0].Short()
	return
}

func (st StackSlice) String() (s string) {
	for _, frame := range st {
		s += "\n" + frame.String()
	}
	return
}

// Format implements fmt.Formatter
func (st StackSlice) Format(s fmt.State, verb rune) {
	if len(st) == 0 {
		return
	}
	switch verb {
	case 'v':
		if s.Flag('-') {
			fmt.Fprint(s, st.Short())
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, st.String())
	case 'q':
		fmt.Fprintf(s, "%q", st.String())
	}
}
