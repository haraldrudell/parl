/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
	"github.com/haraldrudell/parl/pdebug"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// 1 is Go.Register/Go/SubGo/SubGroup/AddError/Done
	// 1 is checkState
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
	goEntityID // Wait()
	// isTerminated indicates that this Go thread is terminated
	//	- an atomic is requires since wg.Done and wg.IsZero are separate operations
	isTerminated    parl.AtomicBool
	goParent        // Cancel() Context()
	creatorThreadId parl.ThreadID
	thread          ThreadSafeThreadData
}

// newGo returns a Go object for a thread operating in a Go thread-group. Thread-safe.
func newGo(parent goParent, goInvocation *pruntime.CodeLocation) (
	g0 parl.Go,
	goEntityID GoEntityID,
	threadData *ThreadData) {
	if parent == nil {
		panic(perrors.NewPF("parent cannot be nil"))
	}
	g := Go{
		goEntityID: *newGoEntityID(),
		goParent:   parent,
	}
	g.wg.Add(1)
	g.creatorThreadId = goid.GoID()
	g.thread.SetCreator(goInvocation)

	g0 = &g
	goEntityID = g.G0ID()
	threadData = g.thread.Get()
	return
}

func (g0 *Go) Register(label ...string) (g00 parl.Go) { return g0.checkState(false, label...) }
func (g0 *Go) Go() (g00 parl.Go)                      { return g0.checkState(false).goParent.FromGoGo() }
func (g0 *Go) SubGo(onFirstFatal ...parl.GoFatalCallback) (subGo parl.SubGo) {
	return g0.checkState(false).goParent.FromGoSubGo(onFirstFatal...)
}
func (g0 *Go) SubGroup(onFirstFatal ...parl.GoFatalCallback) (subGroup parl.SubGroup) {
	return g0.checkState(false).goParent.FromGoSubGroup(onFirstFatal...)
}

func (g0 *Go) AddError(err error) {
	g0.checkState(false)

	if err == nil {
		return // nil error return
	}

	g0.ConsumeError(NewGoError(perrors.Stack(err), parl.GeNonFatal, g0))
}

// Done handles thread exit. Deferrable
func (g0 *Go) Done(errp *error) {
	g0.checkState(true)
	if !g0.isTerminated.Set() {
		panic(perrors.ErrorfPF("Go received multiple Done: ", perrors.ErrpString(errp)))
	}

	// obtain error and ensure it has stack
	var err error
	if errp != nil {
		err = perrors.Stack(*errp)
	}

	g0.goParent.GoDone(g0, err)
	g0.wg.Done()
}

func (g0 *Go) ThreadInfo() (threadData parl.ThreadData) { return g0.thread.Get() }
func (g0 *Go) GoID() (threadID parl.ThreadID)           { return g0.thread.ThreadID() }
func (g0 *Go) Creator() (threadID parl.ThreadID, createLocation *pruntime.CodeLocation) {
	threadID = g0.creatorThreadId
	var threadData = g0.thread.Get()
	createLocation = &threadData.createLocation
	return
}
func (g0 *Go) GoRoutine() (threadID parl.ThreadID, goFunction *pruntime.CodeLocation) {
	var threadData = g0.thread.Get()
	threadID = threadData.threadID
	goFunction = &threadData.funcLocation
	return
}

// checkState is invoked by public methods ensuring that terminated
// objects are not being used
//   - checkState also collects data on the new thread
func (g0 *Go) checkState(skipTerminated bool, label ...string) (g *Go) {
	g = g0
	if !skipTerminated && g0.isTerminated.IsTrue() {
		panic(perrors.NewPF("operation on terminated Go thread object"))
	}

	// ensure we have a threadID
	if g0.thread.HaveThreadID() {
		return
	}

	// update thread information
	var label0 string
	if len(label) > 0 {
		label0 = label[0]
	}
	var stack = pdebug.NewStack(grCheckThreadFrames)
	if stack.IsMain() {
		return // this should not happen, called by Main
	}
	// creator has already been set
	g0.thread.Update(stack.ID(), nil, stack.GoFunction(), label0)

	// propagate thread information
	threadData := g0.thread.Get()
	g0.UpdateThread(g0.G0ID(), threadData)
	return
}

// g1ID:4:g0.(*g1WaitGroup).Go-g1-thread-group.go:63
func (g0 *Go) String() (s string) {
	td := g0.thread.Get()
	return parl.Sprintf("go:%s:%s", td.threadID, td.createLocation.Short())
}
