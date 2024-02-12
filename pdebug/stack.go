/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pdebug provides a portable, parsed stack trace.
package pdebug

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

// - Go stack traces are created by [runtime.NewStack] and is a byte slice
type Stack struct{ pruntime.Stack }

// [pdebug.Stack] implements [parl.Stack]
var _ parl.Stack = &Stack{}

// NewStack populates a Stack object with the current thread
// and its stack using debug.Stack
func NewStack(skipFrames int) (stack parl.Stack) {
	if skipFrames < 0 {
		skipFrames = 0
	}
	// count [pdebug.NewStack]
	skipFrames++

	// result of parsing to be returned
	stack = &Stack{
		Stack: pruntime.NewStack(skipFrames),
	}

	return
}

// thread ID 1… for the thread requesting the stack trace
//   - ThreadID is comparable and has IsValid and String methods
//   - ThreadID is typically an incremented 64-bit integer with
//     main thread having ID 1
func (s *Stack) ID() (threadID parl.ThreadID) {
	threadID = parl.ThreadID(s.Stack.(*pruntime.StackR).ThreadID)
	return
}

// a word indicating thread status, typically word “running”
func (s *Stack) Status() (status parl.ThreadStatus) {
	status = parl.ThreadStatus(s.Stack.(*pruntime.StackR).Status)
	return
}

// the code location of the go statement creating this thread
//   - if IsMain is true, zero-value. Check with Creator().IsSet()
//   - never nil
func (s *Stack) Creator() (creator *pruntime.CodeLocation, creatorID parl.ThreadID, goRoutineRef string) {
	var stackR = s.Stack.(*pruntime.StackR)
	var c = stackR.Creator
	creator = &c
	creatorID = parl.ThreadID(stackR.CreatorID)
	goRoutineRef = stackR.GoroutineRef
	return
}
