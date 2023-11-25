/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/pruntime"
	"golang.org/x/exp/slices"
)

const (
	// 1 is for NewGoGroup/.SubGo/.SubGroup
	//	1 is for new
	goGroupNewObjectFrames = 2
	// 1 is for .Go
	// 1 is for newGo
	goGroupStackFrames  = 2
	goFromGoStackFrames = goGroupStackFrames + 1
	// 1 is for Go method
	// 1 is for NewGoGroup/.SubGo/.SubGroup/.Go
	//	1 is for new
	fromGoNewFrames = goGroupNewObjectFrames + 1
)

// GoGroup is a Go thread-group. Thread-safe.
//   - GoGroup has its own error channel and waitgroup and no parent thread-group.
//   - thread exits are processed by G1Done and the g1WaitGroup
//   - the thread-group terminates when its erropr channel closes
//   - non-fatal erors are processed by ConsumeError and the error channel
//   - new Go threads are handled by the g1WaitGroup
//   - SubGroup creates a subordinate thread-group using this threadgroup’s error channel
type GoGroup struct {
	creator         pruntime.CodeLocation
	parent          goGroupParent
	hasErrorChannel bool // this GoGroup uses its error channel: GoGroup or SubGroup
	isSubGroup      bool // is SubGroup: not GoGroup or SubGo
	onFirstFatal    parl.GoFatalCallback
	gos             parli.ThreadSafeMap[parl.GoEntityID, *ThreadData]
	ch              parl.NBChan[parl.GoError]
	// endCh is a channel that closes when this threadGroup ends
	endCh parl.Awaitable
	// provides Go entity ID, sub-object waitgroup, cancel-context
	*goContext // Cancel() Context() EntityID()

	hadFatal           atomic.Bool
	isNoTermination    atomic.Bool
	isDebug            atomic.Bool
	isAggregateThreads atomic.Bool
	onceWaiter         atomic.Pointer[parl.OnceWaiter]

	// doneLock ensures:
	//	- order of emitted termination goErrors by GoDone
	//	- atomicity of goContext.Done() endCh.Close() ch.Close() in
	//		GoDone Cancel EnableTermination
	//	- ensuring that Add is prevented on GoGroup termination
	//	- doneLock is used in GoDone Add EnableTermination Cancel
	doneLock sync.Mutex // for GoDone method
}

var _ goGroupParent = &GoGroup{}
var _ goParent = &GoGroup{}

// NewGoGroup returns a stand-alone thread-group with its own error channel. Thread-safe.
//   - ctx is not canceled by the thread-group
//   - ctx may initiate thread-group Cancel
//   - a stand-alone GoGroup thread-group has goGroupParent nil
//   - non-fatal and fatal errors from the thread-group’s threads are sent on the GoGroup’s
//     error channel
//   - the GoGroup processes Go invocations and thread-exits from its own threads and
//     the threads of its subordinate thread-groups
//     wait-group and that of its parent
//   - cancel of the GoGroup’s context signals termination to its own threads and all threads of its
//     subordinate thread-groups
//   - the GoGroup’s context is canceled when its provided parent context is canceled or any of its
//     threads invoke the GoGroup’s Cancel method
//   - the GoGroup terminates when its error channel closes from all threads in its own
//     thread-group and that of any subordinate thread-groups have exited.
func NewGoGroup(ctx context.Context, onFirstFatal ...parl.GoFatalCallback) (g0 parl.GoGroup) {
	return new(nil, ctx, true, false, goGroupNewObjectFrames, onFirstFatal...)
}

// Go returns a parl.Go thread-features object
//   - Go is invoked by a g0-package consumer
//   - the Go return value is to be used as a function argument in a go-statement
//     function-call launching a goroutine thread
func (g *GoGroup) Go() (g1 parl.Go) {
	return g.newGo(goGroupStackFrames)
}

// FromGoGo returns a parl.Go thread-features object invoked from another
// parl.Go object
//   - the Go return value is to be used as a function argument in a go-statement
//     function-call launching a goroutine thread
func (g *GoGroup) FromGoGo() (g1 parl.Go) {
	return g.newGo(goFromGoStackFrames)
}

// newGo creates parl.Go objects
func (g *GoGroup) newGo(frames int) (g1 parl.Go) {
	if g.isEnd() {
		panic(perrors.NewPF("after GoGroup termination"))
	}

	// At this point, Go invocation is accessible so retrieve it
	// the goroutine has not been created yet, so there is no creator
	// instead, use top of the stack, the invocation location for the Go() function call
	goInvocation := pruntime.NewCodeLocation(frames)

	// the only location creating Go objects
	var threadData *ThreadData
	var goEntityID parl.GoEntityID
	g1, goEntityID, threadData = newGo(g, goInvocation)

	// count the running thread in this thread-group and its parents
	g.Add(goEntityID, threadData)

	return
}

// newSubGo returns a subordinate thread-group witthout an error channel. Thread-safe.
//   - a SubGo has goGroupParent non-nil and isSubGo true
//   - the SubGo thread’s fatal and non-fatal errors are forwarded to its parent
//   - SubGo has FirstFatal mechanic but no error channel of its own.
//   - the SubGo’s Go invocations and thread-exits are processed by the SubGo’s wait-group
//     and the thread-group of its parent
//   - cancel of the SubGo’s context signals termination to its own threads and all threads of its
//     subordinate thread-groups
//   - the SubGo’s context is canceled when its parent’s context is canceled or any of its
//     threads invoke the SubGo’s Cancel method
//   - the SubGo thread-group terminates when all threads in its own thread-group and
//     that of any subordinate thread-groups have exited.
func (g *GoGroup) SubGo(onFirstFatal ...parl.GoFatalCallback) (g1 parl.SubGo) {
	return new(g, nil, false, false, goGroupNewObjectFrames, onFirstFatal...)
}

// FromGoSubGo returns a subordinate thread-group witthout an error channel. Thread-safe.
func (g *GoGroup) FromGoSubGo(onFirstFatal ...parl.GoFatalCallback) (g1 parl.SubGo) {
	return new(g, nil, false, false, fromGoNewFrames, onFirstFatal...)
}

// newSubGroup returns a subordinate thread-group with an error channel handling fatal
// errors only. Thread-safe.
//   - a SubGroup has goGroupParent non-nil and isSubGo false
//   - fatal errors from the SubGroup’s threads are sent on its own error channel
//   - non-fatal errors from the SubGroup’s threads are forwarded to the parent
//   - the SubGroup’s Go invocations and thread-exits are processed in the SubGroup’s
//     wait-group and that of its parent
//   - cancel of the SubGroup’s context signals termination to its own threads and all threads of its
//     subordinate thread-groups
//   - the SubGroup’s context is canceled when its parent’s context is canceled or any of its
//     threads invoke the SubGroup’s Cancel method
//   - SubGroup thread-group terminates when its error channel closes after all of its threads
//     and threads of its subordinate thread-groups have exited.
func (g *GoGroup) SubGroup(onFirstFatal ...parl.GoFatalCallback) (g1 parl.SubGroup) {
	return new(g, nil, true, true, goGroupNewObjectFrames, onFirstFatal...)
}

// FromGoSubGroup returns a subordinate thread-group with an error channel handling fatal
// errors only. Thread-safe.
func (g *GoGroup) FromGoSubGroup(onFirstFatal ...parl.GoFatalCallback) (g1 parl.SubGroup) {
	return new(g, nil, true, true, fromGoNewFrames, onFirstFatal...)
}

// new returns a new GoGroup as parl.GoGroup
func new(
	parent goGroupParent, ctx context.Context,
	hasErrorChannel, isSubGroup bool,
	stackOffset int,
	onFirstFatal ...parl.GoFatalCallback,
) (g0 *GoGroup) {
	if ctx == nil && parent != nil {
		ctx = parent.Context()
	}
	g := GoGroup{
		creator:   *pruntime.NewCodeLocation(stackOffset),
		parent:    parent,
		goContext: newGoContext(ctx),
		gos:       pmaps.NewRWMap[parl.GoEntityID, *ThreadData](),
		endCh:     *parl.NewAwaitable(),
	}
	if parl.IsThisDebug() {
		g.isDebug.Store(true)
	}
	if len(onFirstFatal) > 0 {
		g.onFirstFatal = onFirstFatal[0]
	}
	if hasErrorChannel {
		g.hasErrorChannel = true
	}
	if isSubGroup {
		g.isSubGroup = true
	}
	if g.isDebug.Load() {
		s := "new:" + g.typeString()
		if parent != nil {
			if p, ok := parent.(*GoGroup); ok {
				s += "(" + p.typeString() + ")"
			}
		}
		parl.Log(s)
	}
	return &g
}

// Add processes a thread from this or a subordinate thread-group
func (g *GoGroup) Add(goEntityID parl.GoEntityID, threadData *ThreadData) {
	g.doneLock.Lock()
	defer g.doneLock.Unlock()

	g.wg.Add(1)
	if g.isDebug.Load() {
		parl.Log("goGroup#%s:Add(id%s:%s)#%d", g.EntityID(), goEntityID, threadData.Short(), g.goContext.wg.Count())
	}
	if g.isAggregateThreads.Load() {
		g.gos.Put(goEntityID, threadData)
	}
	if g.parent != nil {
		g.parent.Add(goEntityID, threadData)
	}
}

// UpdateThread recursively updates thread information for a parl.Go object
// invoked when that Go fiorst obtains the information
func (g *GoGroup) UpdateThread(goEntityID parl.GoEntityID, threadData *ThreadData) {
	if g.isAggregateThreads.Load() {
		g.gos.Put(goEntityID, threadData)
	}
	if g.parent != nil {
		g.parent.UpdateThread(goEntityID, threadData)
	}
}

// Done receives thread exits from threads in subordinate thread-groups
func (g *GoGroup) GoDone(thread parl.Go, err error) {
	if g.endCh.IsClosed() {
		panic(perrors.ErrorfPF("in GoGroup after termination: %s", perrors.Short(err)))
	}

	// first fatal thread-exit of this thread-group
	if err != nil && g.hadFatal.CompareAndSwap(false, true) {

		// handle FirstFatal()
		g.setFirstFatal()

		// onFirstFatal callback
		if g.onFirstFatal != nil {
			var errPanic error
			if errPanic = g.invokeOnFirstFatal(); errPanic != nil {
				g.ConsumeError(NewGoError(
					perrors.ErrorfPF("onFatal callback: %w", errPanic), parl.GeNonFatal, thread))
			}
		}
	}

	// atomic operation: DoneBool and g0.ch.Close
	g.doneLock.Lock()
	defer g.doneLock.Unlock()

	// check inside lock
	if g.endCh.IsClosed() {
		panic(perrors.ErrorfPF("in GoGroup after termination: %s", perrors.Short(err)))
	}

	// debug print termination-start
	if g.isDebug.Load() {
		var threadData parl.ThreadData
		var id string
		if thread != nil {
			threadData = thread.ThreadInfo()
			id = thread.EntityID().String()
		}
		parl.Log("goGroup#%s:GoDone(%sid%s,%s)after#:%d", g.EntityID(), threadData.Short(), id, perrors.Short(err), g.goContext.wg.Count()-1)
	}

	// indicates that this GoGroup is about to terminate
	//	- DoneBool invokes Done and returns status
	var isTermination = g.goContext.wg.DoneBool()

	// delete thread from thread-map
	g.gos.Delete(thread.EntityID(), parli.MapDeleteWithZeroValue)

	// SubGroup with its own error channel with fatals not affecting parent
	//	- send fatal error to parent as non-fatal error with
	//		error context GeLocalChan
	if g.isSubGroup {
		if err != nil {
			g.ConsumeError(NewGoError(err, parl.GeLocalChan, thread))
		}
		// pretend good thread exit to parent
		g.parent.GoDone(thread, nil)
	}

	// emit on local error channel: GoGroup, SubGroup
	if g.hasErrorChannel {
		var goErrorContext parl.GoErrorContext
		if isTermination {
			goErrorContext = parl.GeExit
		} else {
			goErrorContext = parl.GePreDoneExit
		}
		g.ch.Send(NewGoError(err, goErrorContext, thread))
	} else {

		// SubGo: forward error to parent
		g.parent.GoDone(thread, err)
	}

	// debug print termination end
	if g.isDebug.Load() {
		s := "goGroup#" + g.EntityID().String() + ":"
		if isTermination {
			s += parl.Sprintf("Terminated:isSubGroup:%t:hasEc:%t", g.isSubGroup, g.hasErrorChannel)
		} else {
			s += Shorts(g.Threads())
		}
		parl.Log(s)
	}

	if !isTermination {
		return // GoGroup not yet terminated return
	}
	g.endGoGroup()
}

// ConsumeError receives non-fatal errors from a Go thread.
// Go.AddError delegates to this method
func (g *GoGroup) ConsumeError(goError parl.GoError) {
	if g.ch.DidClose() {
		panic(perrors.ErrorfPF("in GoGroup after termination: %s", goError))
	}
	if goError == nil {
		panic(perrors.NewPF("goError cannot be nil"))
	}
	// non-fatal errors are:
	//	- parl.GeNonFatal or
	//	- parl.GeLocalChan when a SubGroup send fatal errors as non-fatal
	if goError.ErrContext() != parl.GeNonFatal && // it is a non-fatal error
		goError.ErrContext() != parl.GeLocalChan { // it is a fatal error store in a local error channel
		panic(perrors.ErrorfPF("G1Group received termination as non-fatal error: goError: %s", goError))
	}

	// it is a non-fatal error that should be processed

	// if we have a parent GoGroup, send it there
	if g.parent != nil {
		g.parent.ConsumeError(goError)
		return
	}

	// send the error to the channel of this stand-alone G1Group
	g.ch.Send(goError)
}

// Ch returns a channel sending the all fatal termination errors when
// the FailChannel option is present, or only the first when both
// FailChannel and StoreSubsequentFail options are present.
func (g *GoGroup) Ch() (ch <-chan parl.GoError) { return g.ch.Ch() }

// FirstFatal allows to await or inspect the first thread terminating with error.
// it is valid if this SubGo has LocalSubGo or LocalChannel options.
// To wait for first fatal error using multiple-semaphore mechanic:
//
//	firstFatal := g0.FirstFatal()
//	for {
//	  select {
//	  case <-firstFatal.Ch():
//	  …
//
// To inspect first fatal:
//
//	if firstFatal.DidOccur() …
func (g *GoGroup) FirstFatal() (firstFatal *parl.OnceWaiterRO) {
	var onceWaiter *parl.OnceWaiter
	for {
		if onceWaiter0 := g.onceWaiter.Load(); onceWaiter0 != nil {
			return parl.NewOnceWaiterRO(onceWaiter0)
		}
		if onceWaiter == nil {
			onceWaiter = parl.NewOnceWaiter(context.Background())
		}
		if g.onceWaiter.CompareAndSwap(nil, onceWaiter) {
			onceWaiter.Cancel()
			return
		}
	}
}

// EnableTermination false prevents the SubGo or GoGroup from terminating
// even if the number of threads is zero
func (g *GoGroup) EnableTermination(allowTermination bool) {
	if g.isDebug.Load() {
		parl.Log("goGroup%s#:EnableTermination:%t", g.EntityID(), allowTermination)
	}
	if g.endCh.IsClosed() {
		return // GoGroup is already shutdown return
	} else if !allowTermination {
		if g.isNoTermination.CompareAndSwap(false, true) { // prevent termination, it was previously allowed
			// add a fake count to parent waitgroup preventing iut from terminating
			g.CascadeEnableTermination(1)
		}
		return // prevent termination complete
	}

	// now allow termination
	if !g.isNoTermination.CompareAndSwap(true, false) {
		return // termination allowed already
	}
	// remove the fake count from parent
	g.CascadeEnableTermination(-1)

	// atomic operation: DoneBool and g0.ch.Close
	g.doneLock.Lock()
	defer g.doneLock.Unlock()

	// if there are suboirdinate objects, termination wwill be done by GoDone
	if !g.wg.IsZero() {
		return // GoGroup is not in pending termination
	}
	g.endGoGroup()
}

// IsEnableTermination returns the state of EnableTermination,
// initially true
func (g *GoGroup) IsEnableTermination() (mayTerminate bool) { return !g.isNoTermination.Load() }

// CascadeEnableTermination manipulates wait groups of this goGroup and
// those of its parents to allow or prevent termination
func (g *GoGroup) CascadeEnableTermination(delta int) {
	g.wg.Add(delta)
	if g.parent != nil {
		g.parent.CascadeEnableTermination(delta)
	}
}

// ThreadsInternal returns values with the internal parl.GoEntityID key
func (g *GoGroup) ThreadsInternal() parli.ThreadSafeMap[parl.GoEntityID, *ThreadData] {
	return g.gos.Clone()
}

// the available data for all threads
func (g *GoGroup) Threads() (threads []parl.ThreadData) {
	// the pointer can be updated at any time, but the value does not change
	list := g.gos.List()
	threads = make([]parl.ThreadData, len(list))
	for i, tp := range list {
		threads[i] = tp
	}
	return
}

// threads that have been named ordered by name
func (g *GoGroup) NamedThreads() (threads []parl.ThreadData) {
	// the pointer can be updated at any time, but the value does not change
	list := g.gos.List()

	// remove unnamed threads
	for i := 0; i < len(list); {
		if list[i].label == "" {
			list = slices.Delete(list, i, i+1)
		} else {
			i++
		}
	}

	// sort pointers
	slices.SortFunc(list, g.cmpNames)

	// return slice of values
	threads = make([]parl.ThreadData, len(list))
	for i, tp := range list {
		threads[i] = tp
	}
	return
}

// SetDebug enables debug logging on this particular instance
//   - parl.NoDebug
//   - parl.DebugPrint
//   - parl.AggregateThread
func (g *GoGroup) SetDebug(debug parl.GoDebug) {
	if debug == parl.DebugPrint {
		g.isDebug.Store(true)
		g.isAggregateThreads.Store(true)
		return
	}
	g.isDebug.Store(false)

	if debug == parl.AggregateThread {
		g.isAggregateThreads.Store(true)
		return
	}

	g.isAggregateThreads.Store(false)
}

// Cancel signals shutdown to all threads of a thread-group.
func (g *GoGroup) Cancel() {

	// cancel the context
	g.goContext.Cancel()

	// check outside lock: done if:
	// - if GoGroup/SubGroup/SubGo already terminated
	//	- subordinate objects exist
	//	- termination is temporarily disabled
	if g.isEnd() || g.goContext.wg.Count() > 0 || g.isNoTermination.Load() {
		return // already ended or have child object or termination off return
	}

	// special case: Cancel before any Go SubGo SubGroup
	//	- normally, GoDone or EnableTermination
	// atomic operation: DoneBool and g0.ch.Close
	g.doneLock.Lock()
	defer g.doneLock.Unlock()

	// repeat check inside lock
	if g.isEnd() || g.goContext.wg.Count() > 0 || g.isNoTermination.Load() {
		return // already ended or have child object or termination off return
	}
	g.endGoGroup()
}

// Wait waits for all threads of this thread-group to terminate.
func (g *GoGroup) Wait() {
	<-g.endCh.Ch()
}

// returns a channel that closes on subGo end similar to Wait
func (g *GoGroup) WaitCh() (ch parl.AwaitableCh) {
	return g.endCh.Ch()
}

// invoked while holding g.doneLock
func (g *GoGroup) endGoGroup() {
	if g.hasErrorChannel {
		g.ch.Close() // close local error channel
	}
	// mark GoGroup terminated
	g.endCh.Close()
	// cancel the context
	g.goContext.Cancel()
}

// cmpNames is a slice comparison function for thread names
func (g *GoGroup) cmpNames(a *ThreadData, b *ThreadData) (result int) {
	if a.label < b.label {
		return -1
	} else if a.label > b.label {
		return 1
	}
	return 0
}

// setFirstFatal triggers a possible onFirstFatal
func (g *GoGroup) setFirstFatal() {
	var onceWaiter = g.onceWaiter.Load()
	if onceWaiter == nil {
		return // FirstFatal not invoked return
	}
	onceWaiter.Cancel()
}

// isEnd determines if this goGroup has ended
//   - if goGroup has error channel, the goGroup ends when its error channel closes
//   - — goGroups without a parent
//   - — subGroups with error channel
//   - — a subGo, having no error channel, ends when all threads have exited
//   - if the GoGroup or any of its subordinate thread-groups have EnableTermination false
//     GoGroups will not end until EnableTermination true
func (g *GoGroup) isEnd() (isEnd bool) {

	// SubGo termination flag
	if !g.hasErrorChannel {
		return g.endCh.IsClosed()
	}

	// others is when error channel closes
	var ch = g.ch.Ch()
	select {
	case <-ch:
		isEnd = true
	default:
	}

	return
}

// "goGroup#1" "subGroup#2" "subGo#3"
func (g *GoGroup) typeString() (s string) {
	if g.parent == nil {
		s = "goGroup"
	} else if g.isSubGroup {
		s = "subGroup"
	} else {
		s = "subGo"
	}
	return s + "#" + g.goEntityID.EntityID().String()
}

// g1Group#3threads:1(1)g0.TestNewG1Group-g1-group_test.go:60
func (g *GoGroup) String() (s string) {
	return parl.Sprintf("%s_threads:%s_New:%s",
		g.typeString(), // "goGroup#1"
		g.goContext.wg.String(),
		g.creator.Short(),
	)
}

func (g *GoGroup) invokeOnFirstFatal() (err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	g.onFirstFatal(g)

	return
}
