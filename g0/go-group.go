/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"fmt"
	"strings"
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
	// creator is the code line that invoked new for this GoGroup SubGo or SubGroup
	creator pruntime.CodeLocation
	// parent for SubGo SubGroup, nil for GoGroup
	parent goGroupParent
	// true if instance has error channel, ie. is GoGroup or SubGroup
	hasErrorChannel bool
	// true if instance is SubGroup and not GoGroup or SubGo
	isSubGroup bool
	// invoked on first fatal thread-exit
	onFirstFatal parl.GoFatalCallback
	// gos is a map from goEntityId to subordinate SubGo SunGroup Go
	gos parli.ThreadSafeMap[parl.GoEntityID, *ThreadData]
	// unbound error channel used when instance is GoGroup or SubGroup
	errCh parl.NBChan[parl.GoError]
	// channel that closes when this threadGroup ends
	endCh parl.Awaitable
	// provides Go entity ID, sub-object waitgroup, cancel-context
	//	- Cancel() Context() EntityID()
	goContext

	// whether a fatal exit has occurred
	hadFatal atomic.Bool
	// whether thread-group termination is allowed
	//	- set by EnableTermination
	isNoTermination atomic.Bool
	// controls whether debug information is printed
	//	- set by SetDebug
	isDebug atomic.Bool
	// controls whether trean information is stroed in gos
	//	- set by SetDebug
	isAggregateThreads atomic.Bool
	onceWaiter         atomic.Pointer[parl.OnceWaiter]
	// debug-log set by SetDebug
	log atomic.Pointer[parl.PrintfFunc]

	// doneLock ensures:
	//	- critical section for:
	//	- — closing of error channel
	//	- — change of number of child objects or ending that waitGroup
	//	- — change in enableTermination state
	//	- order of:
	//	- — parent Add GoDone
	//	- — emitted termination-goErrors by [GoGroup.GoDone]
	//	- mutual exclusion of:
	//	- — [GoGroup.GoDone]
	//	- — [GoGroup.Cancel]
	//	- — [GoGroup.Add]
	//	- — [GoGroup.EnableTermination]
	//	- the context can be canceled at any time
	doneLock sync.Mutex
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
func (g *GoGroup) Go() (g2 parl.Go) { return g.newGo(goGroupStackFrames) }

// FromGoGo returns a parl.Go thread-features object invoked from another
// parl.Go object
//   - the Go return value is to be used as a function argument in a go-statement
//     function-call launching a goroutine thread
func (g *GoGroup) FromGoGo() (g2 parl.Go) { return g.newGo(goFromGoStackFrames) }

// newGo creates parl.Go objects
func (g *GoGroup) newGo(frames int) (g2 parl.Go) {
	// At this point, Go invocation is accessible so retrieve it
	// the goroutine has not been created yet, so there is no creator
	// instead, use top of the stack, the invocation location for the Go() function call
	var goInvocation = pruntime.NewCodeLocation(frames)

	if g.isEnd() {
		panic(perrors.ErrorfPF(g.panicString(".Go(): "+goInvocation.Short(), nil, nil, false, nil)))
	}

	// the only location creating Go objects
	var threadData *ThreadData
	var goEntityID parl.GoEntityID
	g2, goEntityID, threadData = newGo(g, goInvocation)

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
func (g *GoGroup) SubGo(onFirstFatal ...parl.GoFatalCallback) (g2 parl.SubGo) {
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
func (g *GoGroup) SubGroup(onFirstFatal ...parl.GoFatalCallback) (g2 parl.SubGroup) {
	return new(g, nil, true, true, goGroupNewObjectFrames, onFirstFatal...)
}

// FromGoSubGroup returns a subordinate thread-group with an error channel handling fatal
// errors only. Thread-safe.
func (g *GoGroup) FromGoSubGroup(onFirstFatal ...parl.GoFatalCallback) (g2 parl.SubGroup) {
	return new(g, nil, true, true, fromGoNewFrames, onFirstFatal...)
}

// new returns a new GoGroup as parl.GoGroup
func new(
	parent goGroupParent, ctx context.Context,
	hasErrorChannel, isSubGroup bool,
	stackOffset int,
	onFirstFatal ...parl.GoFatalCallback,
) (g2 *GoGroup) {
	if ctx == nil && parent != nil {
		ctx = parent.Context()
	}
	g := GoGroup{
		creator: *pruntime.NewCodeLocation(stackOffset),
		parent:  parent,
		gos:     pmaps.NewRWMap[parl.GoEntityID, *ThreadData](),
	}
	newGoContext(&g.goContext, ctx)
	if parl.IsThisDebug() {
		g.isDebug.Store(true)
		var log parl.PrintfFunc = parl.Log
		g.log.CompareAndSwap(nil, &log)
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
		(*g.log.Load())(s)
	}
	return &g
}

// Add processes a thread from this or a subordinate thread-group
func (g *GoGroup) Add(goEntityID parl.GoEntityID, threadData *ThreadData) {
	g.doneLock.Lock() // Add
	defer g.doneLock.Unlock()

	g.wg.Add(1)
	if g.isDebug.Load() {
		(*g.log.Load())("goGroup#%s:Add(new:Go#%s.Go():%s)#%d",
			g.EntityID(),
			goEntityID, threadData.Short(), g.goContext.wg.Count())
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
		panic(perrors.ErrorfPF(g.panicString("", thread, &err, false, nil)))
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
	g.doneLock.Lock() // GoDone
	defer g.doneLock.Unlock()

	// check inside lock
	if g.endCh.IsClosed() {
		panic(perrors.ErrorfPF(g.panicString("", thread, &err, false, nil)))
	}

	// debug print termination-start
	if g.isDebug.Load() {
		var threadData parl.ThreadData
		var id string
		if thread != nil {
			threadData = thread.ThreadInfo()
			id = thread.EntityID().String()
		}
		(*g.log.Load())("goGroup#%s:GoDone(Label-ThreadID:%sGo#%s_exit:‘%s’)after#:%d",
			g.EntityID(),
			threadData.Short(), id, perrors.Short(err),
			g.goContext.wg.Count()-1,
		)
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
		g.errCh.Send(NewGoError(err, goErrorContext, thread))
	} else {

		// SubGo: forward error to parent
		g.parent.GoDone(thread, err)
	}

	// debug print termination end
	if g.isDebug.Load() {
		var actionS string
		if isTermination {
			actionS = fmt.Sprintf("TERMINATED:isSubGroup:%t:hasEc:%t", g.isSubGroup, g.hasErrorChannel)
		} else {
			actionS = "remaining:" + Shorts(g.Threads())
		}
		(*g.log.Load())(fmt.Sprintf("%s:%s", g.typeString(), actionS))
	}

	if !isTermination {
		return // GoGroup not yet terminated return
	}
	g.endGoGroup() // GoDone
}

// ConsumeError receives non-fatal errors from a Go thread.
//   - Go.AddError delegates to this method
func (g *GoGroup) ConsumeError(goError parl.GoError) {
	if g.errCh.DidClose() {
		panic(perrors.ErrorfPF(g.panicString("", nil, nil, true, goError)))
	}
	if goError == nil {
		panic(perrors.NewPF("goError cannot be nil"))
	}
	// non-fatal errors are:
	//	- parl.GeNonFatal or
	//	- parl.GeLocalChan when a SubGroup send fatal errors as non-fatal
	if goError.ErrContext() != parl.GeNonFatal && // it is a non-fatal error
		goError.ErrContext() != parl.GeLocalChan { // it is a fatal error store in a local error channel
		panic(perrors.ErrorfPF(g.panicString("received termination as non-fatal error", nil, nil, true, goError)))
	}

	// it is a non-fatal error that should be processed

	// if we have a parent GoGroup, send it there
	if g.parent != nil {
		g.parent.ConsumeError(goError)
		return
	}

	// send the error to the channel of this stand-alone G1Group
	g.errCh.Send(goError)
}

// Ch returns a channel sending the all fatal termination errors when
// the FailChannel option is present, or only the first when both
// FailChannel and StoreSubsequentFail options are present.
func (g *GoGroup) Ch() (ch <-chan parl.GoError) { return g.errCh.Ch() }

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

// EnableTermination controls whether the thread-droup is allowed to terminate
//   - true is default
//   - period of false prevents terminating even if child-object count reaches zero
//   - invoking with true while child-object count is zero,
//     terminates the thread-group regardless of previous enableTermination state.
//     This is used prior to Wait when a thread-group was not used.
//     Using the alternative Cancel would signal to threads to exit.
func (g *GoGroup) EnableTermination(allowTermination ...bool) (mayTerminate bool) {
	if g.isDebug.Load() {
		(*g.log.Load())("%s:EnableTermination:%t", g.typeString(), allowTermination)
	}

	// if no argument or the thread-group already terminated
	//	- just return current state
	if len(allowTermination) == 0 || g.endCh.IsClosed() {
		return !g.isNoTermination.Load()
	}

	// prevent termination case
	if !allowTermination[0] {
		// cascade if it is a state change
		if g.isNoTermination.CompareAndSwap(false, true) {
			// add a fake count to parent waitgroup preventing iut from terminating
			g.CascadeEnableTermination(1)
		}
		return // prevent termination complete: mayTerminate: false
	}

	// allow termination case
	//	- must always be cascaded
	//	- either to change the state or
	//	- for unused thread-group termination
	var delta int
	if g.isNoTermination.Load() {
		if g.isNoTermination.CompareAndSwap(true, false) {
			// remove the fake count from parent
			delta = -1
			g.CascadeEnableTermination(delta)
		}
	}
	if delta == 0 {
		if p := g.parent; p != nil {
			p.CascadeEnableTermination(0)
		}
	}

	mayTerminate = true

	// check if this thread-group should be terminated
	// atomic operation: DoneBool and g0.ch.Close
	g.doneLock.Lock() // EnableTermination
	defer g.doneLock.Unlock()

	// if there are subordinate objects, termination will be done by GoDone
	if !g.wg.IsZero() {
		return // GoGroup is not in pending termination
	}
	// all threads have exited, so this ends the thread-group
	if g.isDebug.Load() {
		(*g.log.Load())("%s:TERMINATED:EnableTermination", g.typeString())
	}
	g.endGoGroup() // EnableTermination

	return
}

// CascadeEnableTermination manipulates wait groups of this goGroup and
// those of its parents to allow or prevent termination
func (g *GoGroup) CascadeEnableTermination(delta int) {
	g.wg.Add(delta)
	if g.parent != nil {
		g.parent.CascadeEnableTermination(delta)
	}
	// make EnableTermination Allow cascade
	if delta == 0 && !g.isNoTermination.Load() {
		g.EnableTermination(parl.AllowTermination)
	}
}

// ThreadsInternal returns values with the internal parl.GoEntityID key
func (g *GoGroup) ThreadsInternal() (m parli.ThreadSafeMap[parl.GoEntityID, *ThreadData]) {
	return g.gos.Clone()
}

// Internals returns methods used by [g0debug.ThreadLogger]
func (g *GoGroup) Internals() (
	isEnd func() bool,
	isAggregateThreads *atomic.Bool,
	setCancelListener func(f func()),
	endCh <-chan struct{},
) {
	if g.hasErrorChannel {
		endCh = g.errCh.WaitForCloseCh()
	} else {
		endCh = g.endCh.Ch()
	}
	return g.isEnd, &g.isAggregateThreads, g.goContext.setCancelListener, endCh
}

// the available data for all threads
func (g *GoGroup) Threads() (threads []parl.ThreadData) {
	// the pointer can be updated at any time, but the value does not change
	var list = g.gos.List()
	threads = make([]parl.ThreadData, len(list))
	for i, tp := range list {
		threads[i] = tp
	}
	return
}

// threads that have been named ordered by name
func (g *GoGroup) NamedThreads() (threads []parl.ThreadData) {
	// the pointer can be updated at any time, but the value does not change
	//	- slice of struct pointer
	var list = g.gos.List()

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

	// return slice of interface
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
func (g *GoGroup) SetDebug(debug parl.GoDebug, log ...parl.PrintfFunc) {

	// ensure g.log
	var logF parl.PrintfFunc
	if len(log) > 0 {
		logF = log[0]
	}
	if logF != nil {
		g.log.Store(&logF)
	} else if g.log.Load() == nil {
		logF = parl.Log
		g.log.Store(&logF)
	}

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
	g.doneLock.Lock() // Cancel
	defer g.doneLock.Unlock()

	// repeat check inside lock
	if g.isEnd() || g.goContext.wg.Count() > 0 || g.isNoTermination.Load() {
		return // already ended or have child objects or termination off return
	}
	if g.isDebug.Load() {
		(*g.log.Load())("%s:TERMINATED:Cancel", g.typeString())
	}
	g.endGoGroup() // Cancel
}

// Wait waits for all threads of this thread-group to terminate.
func (g *GoGroup) Wait() {
	<-g.endCh.Ch()
}

// returns a channel that closes on subGo end similar to Wait
func (g *GoGroup) WaitCh() (ch parl.AwaitableCh) {
	return g.endCh.Ch()
}

func (g *GoGroup) panicString(
	text string,
	thread parl.Go,
	errp *error,
	hasGoE bool, goError parl.GoError,
) (s string) {
	var sL = []string{fmt.Sprintf("after %s termination.", g.typeString())}
	if text != "" {
		sL = append(sL, text)
	}
	if thread != nil {
		var _, goFunction = thread.GoRoutine()
		if goFunction.IsSet() {
			sL = append(sL, "goFunc: "+goFunction.Short())
		} else {
			var _, creator = thread.Creator()
			sL = append(sL, "go-statement: "+creator.Short())
		}
	}
	if errp != nil {
		sL = append(sL, fmt.Sprintf("err: ‘%s’", perrors.Short(*errp)))
	}
	if hasGoE {
		sL = append(sL, goError.String())
	}
	sL = append(sL, "newGroup: "+g.creator.Short())
	return strings.Join(sL, "\x20")
}

// invoked while holding g.doneLock
//   - closes error channel if GoGroup or SubGroup
//   - closes endCh
//   - cancels context
func (g *GoGroup) endGoGroup() {
	if g.hasErrorChannel {
		g.errCh.Close() // close local error channel
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
	} else {
		// others is when error channel closes
		return g.errCh.IsClosed()
	}
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
