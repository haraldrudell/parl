/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"fmt"
	"io"
	"runtime"
)

const (
	maxStackFrameSize = 1024
	// 0 is runtime.Callers
	// 1 is NewStackSLice
	// 2 is caller location
	newStackSliceFramesToSkip = 2
)

// StackSlice represents a StackSlice of program counters.
type StackSlice []CodeLocation

// NewStackSlice gets a slice of stack frames
func NewStackSlice(skip int) (slice StackSlice) {

	// get the slice of runtime.Frames
	var frames *runtime.Frames
	for pcs := make([]uintptr, maxStackFrameSize); ; pcs = make([]uintptr, 2*len(pcs)) {
		if entries := runtime.Callers(newStackSliceFramesToSkip+skip, pcs); entries < len(pcs) {
			frames = runtime.CallersFrames(pcs[:entries])
			break // the stack fit into pcs slice
		}
	}

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
	if len(st) >= 1 {
		s = " at " + st[0].Short()
	}
	return
}

func (st StackSlice) Clone() (s StackSlice) {
	s = make([]CodeLocation, len(st))
	copy(s, st)
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
