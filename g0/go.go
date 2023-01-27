/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pdebug"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	goFrames            = 1 // newG1 is invoked from interrnal Go() function
	grCheckThreadFrames = 0
)

// Go supports a goroutine executing part of a thread-group. Thread-safe.
//   - a single thread exit is handled by Done, delegated to parent, and is deferrable.
//   - the Go thread terminates on Done invocation.
//   - non-fatal errors are processed by AddError delegated to parent
//   - new Go threads are delegated to parent
//   - SubGo creates a subordinate thread-group with its own error channel
//   - SubGroup creates a subordinate thread-group with FirstFatal mechanic
type Go struct {
	goEntityID
	isTerminated parl.AtomicBool
	goParent     // ConsumeError() Go() Cancel() Context()
	thread       ThreadSafeThreadData
	goWaitGroup
}

// newGo returns a Go object for a thread operating in a Go thread-group. Thread-safe.
func newGo(parent goParentArg, goInvocation *pruntime.CodeLocation) (
	g0 parl.Go,
	goEntityID GoEntityID,
	threadData *ThreadData) {
	if parent == nil {
		panic(perrors.NewPF("parent cannot be nil"))
	}
	g := Go{
		goEntityID:  *newGoEntityID(goFrames),
		goParent:    parent,
		goWaitGroup: *newGoWaitGroup(parent.Context()),
	}
	g.wg.Add(1)
	g.thread.SetCreator(goInvocation)

	goEntityID = g.G0ID()
	threadData = g.thread.Get()
	g0 = &g
	return
}

func (g0 *Go) Register(label ...string) (g00 parl.Go) { g0.checkState(false, label...); return g0 }

// SubGo returns a thread-group without its own error channel but
// with FirstFatal mechanic
func (g0 *Go) SubGo(onFirstFatal ...parl.GoFatalCallback) (subGo parl.SubGo) {
	g0.checkState(false)
	return g0.goParent.SubGo(onFirstFatal...)
}

// SubGroup returns a thread-group with its own error channel.
func (g0 *Go) SubGroup(onFirstFatal ...parl.GoFatalCallback) (subGroup parl.SubGroup) {
	g0.checkState(false)
	return g0.goParent.SubGroup(onFirstFatal...)
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

	g0.goContext.Cancel()
	g0.goParent.GoDone(g0, err)
}

func (g0 *Go) ThreadData() (threadData *ThreadData) {
	threadData = g0.thread.Get()
	return
}

// Wait awaits exit of this Go thread
func (g0 *Go) Wait() {
	g0.wg.Wait()
}

// CancelGo signals to this Go thread to exit.
func (g0 *Go) CancelGo() {
	g0.goWaitGroup.Cancel()
}

// Cancel cancels the GoGroup
func (g0 *Go) Cancel() {
	g0.goParent.Cancel()
}

func (g0 *Go) ThreadInfo() (threadData parl.ThreadData) {
	threadData = g0.thread.Get() // a copy of data extracted from behind lock
	return
}

// checkState is invoked by all public methods ensuring that terminated
// objects are not being used
//   - checkState also collects data on the new thread
func (g0 *Go) checkState(skipTerminated bool, label ...string) {
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
	g0.thread.Update(pdebug.NewStack(grCheckThreadFrames), label0)

	// propagate thread information
	threadData := g0.thread.Get()
	g0.UpdateThread(g0.G0ID(), threadData)
}

// g1ID:4:g0.(*g1WaitGroup).Go-g1-thread-group.go:63
func (g0 *Go) String() (s string) {
	td := g0.thread.Get()
	return parl.Sprintf("go:%s:%s", td.threadID, td.createLocation.Short())
}
