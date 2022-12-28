/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

// ThreadData contains identifiable information about a running thread.
//   - ThreadData does not have initialization
type ThreadData struct {
	// threadID is the ID of the running thread
	threadID parl.ThreadID
	// createLocation is the code line of the go-statement function-call
	// invocation launching the thread
	createLocation pruntime.CodeLocation
	// funcLocation is the code line of the function of the running thread.
	funcLocation pruntime.CodeLocation
}

// Update populates this object from a stack trace.
func (td *ThreadData) Update(stack parl.Stack) {
	td.threadID = stack.ID()
	td.createLocation = *stack.Creator()
	td.funcLocation = *stack.Frames()[len(stack.Frames())-1].Loc()
}

// SetCreator gets preliminary Go identifier: the line invoking Go()
func (td *ThreadData) SetCreator(cl *pruntime.CodeLocation) {
	td.createLocation = *cl
}

// threadID is the ID of the running thread
func (td *ThreadData) ThreadID() (threadID parl.ThreadID) {
	return td.threadID
}

func (td *ThreadData) Get() (threadID parl.ThreadID, createLocation pruntime.CodeLocation,
	funcLocation pruntime.CodeLocation) {
	threadID = td.threadID
	createLocation = td.createLocation
	funcLocation = td.funcLocation
	return
}
