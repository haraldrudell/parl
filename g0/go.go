/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package g0 provides Go threads and thread-groups
package g0

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
	"github.com/haraldrudell/parl/pdebug"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// counts public method: Go.Register/Go/SubGo/SubGroup/AddError/Done
	// and [Go.checkState]
	grCheckThreadFrames = 2
)

// Go supports a goroutine executing part of a thread-group. Thread-safe.
//   - Register allows to name the thread and collects information on the new thread
//   - AddError processes non-fatal errors
//   - Done handles thread termination and fatal errors and is deferrable.
//     The Go thread terminates on Done invocation.
//   - Go creates a sibling thread object to be provided in a go statement
//   - SubGo creates a subordinate thread-group managed by the parent thread group.
//     The SubGo can be independently terminated or terminated prior to thread exit.
//   - SubGroup creates a subordinate thread-group with its own error channel.
//     Fatal-error thread-exits in SubGroup can be recovered locally in that thread-group
type Go struct {
	// EntityID(), uniquely identifies Go object
	goEntityID
	// Cancel() Context()
	goParent
	// the thread ID of the goroutine creating this thread
	creatorThreadId parl.ThreadID
	// this thread’s Thread ID, creator location, go-function and
	// possible printable thread-name
	thread *ThreadSafeThreadData
	// [parl.AwaitableCh] that closes when this Go ends
	endCh parl.Awaitable
}

// newGo returns a Go object providing functions to a thread operating in a
// Go thread-group. Thread-safe
//   - parent is a GoGroup type configured as GoGroup, SubGo or SubGroup
//   - goInvocation is the invoker of Go, ie. the parent thread
//   - returns Go entity ID and thread data since the parent will
//     immediately need those
func newGo(parent goParent, goInvocation *pruntime.CodeLocation) (
	g0 parl.Go,
	goEntityID parl.GoEntityID,
	threadData *ThreadData) {
	if parent == nil {
		panic(perrors.NewPF("parent cannot be nil"))
	}
	g := Go{
		goEntityID:      *newGoEntityID(),
		goParent:        parent,
		creatorThreadId: goid.GoID(),
		thread:          NewThreadSafeThreadData(),
	}
	g.thread.SetCreator(goInvocation)

	// return values
	g0 = &g
	goEntityID = g.EntityID()
	threadData = g.thread.Get()

	return
}

func (g *Go) Register(label ...string) (g00 parl.Go) { return g.ensureThreadData(label...) }
func (g *Go) Go() (g00 parl.Go)                      { return g.ensureThreadData().goParent.FromGoGo() }
func (g *Go) SubGo(onFirstFatal ...parl.GoFatalCallback) (subGo parl.SubGo) {
	return g.ensureThreadData().goParent.FromGoSubGo(onFirstFatal...)
}
func (g *Go) SubGroup(onFirstFatal ...parl.GoFatalCallback) (subGroup parl.SubGroup) {
	return g.ensureThreadData().goParent.FromGoSubGroup(onFirstFatal...)
}

// AddError emits a non-fatal errors
func (g *Go) AddError(err error) {
	g.ensureThreadData()

	if err == nil {
		return // nil error return
	}

	g.ConsumeError(NewGoError(perrors.Stack(err), parl.GeNonFatal, g))
}

// Done handles thread exit. Deferrable
//   - *errp contains possible fatalk thread error
//   - errp can be nil
func (g *Go) Done(errp *error) {
	if !g.ensureThreadData().endCh.Close() {
		panic(perrors.ErrorfPF("Go received multiple Done: ", perrors.ErrpString(errp)))
	}

	// obtain fatal error and ensure it has stack
	var err error
	if errp != nil {
		err = perrors.Stack(*errp)
	}

	// notify parent of exit
	g.goParent.GoDone(g, err)
}

func (g *Go) ThreadInfo() (threadData parl.ThreadData) { return g.thread.Get() }
func (g *Go) GoID() (threadID parl.ThreadID)           { return g.thread.ThreadID() }
func (g *Go) Creator() (threadID parl.ThreadID, createLocation *pruntime.CodeLocation) {
	threadID = g.creatorThreadId
	var threadData = g.thread.Get()
	createLocation = &threadData.createLocation
	return
}
func (g *Go) GoRoutine() (threadID parl.ThreadID, goFunction *pruntime.CodeLocation) {
	var threadData = g.thread.Get()
	threadID = threadData.threadID
	goFunction = &threadData.funcLocation
	return
}
func (g *Go) Wait()                         { <-g.endCh.Ch() }
func (g *Go) WaitCh() (ch parl.AwaitableCh) { return g.endCh.Ch() }

// ensureThreadData is invoked by Go’s public methods ensuring that
// the thread’s information is collected
//   - label is an optional printable thread-name
//   - ensureThreadData supports functional chaining
func (g *Go) ensureThreadData(label ...string) (g1 *Go) {
	g1 = g

	// if thread-data has already been collected, do nothing
	if g.thread.HaveThreadID() {
		return // already have thread-data return
	}

	// optional printable thread name
	var label0 string
	if len(label) > 0 {
		label0 = label[0]
	}

	// get stack that contains thread ID, go function, go-function invoker
	// for the new thread
	var stack = pdebug.NewStack(grCheckThreadFrames)
	if stack.IsMain() {
		return // this should not happen, called by Main
	}
	// creator has already been set
	g.thread.Update(stack.ID(), nil, stack.GoFunction(), label0)

	// propagate thread information to parent
	g.UpdateThread(g.EntityID(), g.thread.Get())

	return
}

// g1ID:4:g0.(*g1WaitGroup).Go-g1-thread-group.go:63
func (g *Go) String() (s string) {
	td := g.thread.Get()
	return parl.Sprintf("go:%s:%s", td.threadID, td.createLocation.Short())
}
