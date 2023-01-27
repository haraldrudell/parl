/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"strconv"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/pruntime"
	"golang.org/x/exp/slices"
)

const (
	goGroupExtraFrames = 0
)

// GoGroup is a Go thread-group. Thread-safe.
//   - GoGroup has its own error channel and waitgroup and no parent thread-group.
//   - thread exits are processed by G1Done and the g1WaitGroup
//   - the thread-group terminates when its erropr channel closes
//   - non-fatal erors are processed by ConsumeError and the error channel
//   - new Go threads are handled by the g1WaitGroup
//   - SubGroup creates a subordinate thread-group using this threadgroup’s error channel
type GoGroup struct {
	goEntityID
	parent          goGroupParent
	hasErrorChannel parl.AtomicBool // this GoGroup uses its error channel: NewGoGroup() or SubGroup()
	isSubGroup      parl.AtomicBool // is SubGroup(): not NewGoGroup() or SubGo()
	isWaitGroupDone parl.AtomicBool
	hadFatal        parl.AtomicBool
	onFirstFatal    parl.GoFatalCallback
	gos             parli.ThreadSafeMap[GoEntityID, *ThreadData]
	goWaitGroup     // Go() SubGroup() Wait()
	ch              parl.NBChan[parl.GoError]
	doneLock        sync.Mutex // for GoDone method
	noTermination   parl.AtomicBool
	fakeGo          parl.AtomicReference[Go]
	isDebug         parl.AtomicBool

	owLock     sync.Mutex
	onceWaiter *parl.OnceWaiter
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
	return new(nil, ctx, true, false, onFirstFatal...)
}

// Go returns a parl.Go thread-features object
//   - Go is invoked by a g0-package consumer
//   - the Go return value is to be used as a function argument in a go-statement
//     function-call launching a goroutine thread
func (g0 *GoGroup) Go() (g1 parl.Go) {
	if g0.isEnd() {
		panic(perrors.NewPF("after GoGroup termination"))
	}

	// At this point, Go invocation is accessible so retrieve it
	// the goroutine has not been created yet, so there is no creator
	// instead, use top of the stack, the invocation location for the Go() function call
	goInvocation := pruntime.NewCodeLocation(g1Gropupg0StackFranmes)

	// the only location creating Go objects
	var threadData *ThreadData
	var goEntityID GoEntityID
	g1, goEntityID, threadData = newGo(g0, goInvocation)

	// count the running thread in this thread-group and its parents
	g0.Add(goEntityID, threadData)

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
func (g0 *GoGroup) SubGo(onFirstFatal ...parl.GoFatalCallback) (g1 parl.SubGo) {
	return new(g0, nil, false, false, onFirstFatal...)
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
func (g0 *GoGroup) SubGroup(onFirstFatal ...parl.GoFatalCallback) (g1 parl.SubGroup) {
	return new(g0, nil, true, true, onFirstFatal...)
}

// new returns a new GoGroup as parl.GoGroup
func new(
	parent goGroupParent, ctx context.Context,
	hasErrorChannel, isSubGroup bool,
	onFirstFatal ...parl.GoFatalCallback,
) (g0 *GoGroup) {
	if ctx == nil && parent != nil {
		ctx = parent.Context()
	}
	g := GoGroup{
		goEntityID:  *newGoEntityID(goGroupExtraFrames),
		parent:      parent,
		gos:         pmaps.NewRWMap[GoEntityID, *ThreadData](),
		goWaitGroup: *newGoWaitGroup(ctx),
	}
	if parl.IsThisDebug() {
		g.isDebug.Set()
	}
	if len(onFirstFatal) > 0 {
		g.onFirstFatal = onFirstFatal[0]
	}
	if hasErrorChannel {
		g.hasErrorChannel.Set()
	}
	if isSubGroup {
		g.isSubGroup.Set()
	}
	if g.isDebug.IsTrue() {
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
func (g0 *GoGroup) Add(goEntityID GoEntityID, threadData *ThreadData) {
	g0.doneLock.Lock()
	defer g0.doneLock.Unlock()

	g0.wg.Add(1)
	if g0.isDebug.IsTrue() {
		parl.Log("goGroup#%s:Add(id%s:%s)#%d", g0.G0ID(), goEntityID, threadData.Short(), g0.goWaitGroup.wg.Count())
	}
	g0.gos.Put(goEntityID, threadData)
	if g0.parent != nil {
		g0.parent.Add(goEntityID, threadData)
	}
}

func (g0 *GoGroup) UpdateThread(goEntityID GoEntityID, threadData *ThreadData) {
	g0.gos.Put(goEntityID, threadData)
	if g0.parent != nil {
		g0.parent.UpdateThread(goEntityID, threadData)
	}
}

// Done receives thread exits from threads in subordinate thread-groups
func (g0 *GoGroup) GoDone(thread parl.Go, err error) {
	if g0.isWaitGroupDone.IsTrue() {
		panic(perrors.ErrorfPF("in GoGroup after termination: %s", perrors.Short(err)))
	}

	// first fatal thread-exit of this thread-group
	if err != nil && g0.hadFatal.Set() {

		// handle FirstFatal()
		g0.setFirstFatal()

		// onFirstFatal callback
		if g0.onFirstFatal != nil {
			var errPanic error
			parl.RecoverInvocationPanic(func() {
				g0.onFirstFatal(g0)
			}, &errPanic)
			if errPanic != nil {
				g0.ConsumeError(NewGoError(
					perrors.ErrorfPF("onFatal callback: %w", errPanic), parl.GeNonFatal, thread))
			}
		}
	}

	// atomic operation: DoneBool and g0.ch.Close
	g0.doneLock.Lock()
	defer g0.doneLock.Unlock()

	if g0.isDebug.IsTrue() {
		var threadData parl.ThreadData
		var id string
		if thread != nil {
			threadData = thread.ThreadInfo()
			id = thread.(*Go).G0ID().String()
		}
		parl.Log("goGroup#%s:GoDone(%sid%s,%s)after#:%d", g0.G0ID(), threadData.Short(), id, perrors.Short(err), g0.goWaitGroup.wg.Count()-1)
	}

	// process thread-exit
	isTermination := g0.goWaitGroup.wg.DoneBool()
	var threadGo goImpl
	var ok bool
	if threadGo, ok = thread.(*Go); !ok {
		panic(perrors.NewPF("type assertion failed"))
	}
	g0.gos.Delete(threadGo.G0ID())
	if g0.isSubGroup.IsTrue() {

		// SubGroup with its own error channel with fatals not affecting parent
		// send fatal error to parent as non-fatal error with error context GeLocalChan
		if err != nil {
			g0.ConsumeError(NewGoError(err, parl.GeLocalChan, thread))
		}
		// pretend good thread exit to parent
		g0.parent.GoDone(thread, nil)
	}
	if g0.hasErrorChannel.IsTrue() {

		// emit on local error channel
		var context parl.GoErrorContext
		if isTermination {
			context = parl.GeExit
		} else {
			context = parl.GePreDoneExit
		}
		g0.ch.Send(NewGoError(err, context, thread))
		if isTermination {
			g0.ch.Close() // close local error channel
		}
	} else {

		// SubGo case: all forwarded to parent
		g0.parent.GoDone(thread, err)
	}

	if g0.isDebug.IsTrue() {
		s := "goGroup#" + g0.G0ID().String() + ":"
		if isTermination {
			s += parl.Sprintf("Terminated:isSubGroup:%t:hasEc:%t", g0.isSubGroup.IsTrue(), g0.hasErrorChannel.IsTrue())
		} else {
			s += Shorts(g0.Threads())
		}
		parl.Log(s)
	}

	if !isTermination {
		return // GoGroup not yet terminated return
	}

	// mark GoGroup terminated
	g0.isWaitGroupDone.Set()
	g0.goContext.Cancel()
}

// ConsumeError receives non-fatal errors from a Go thread.
// Go.AddError delegates to this method
func (g0 *GoGroup) ConsumeError(goError parl.GoError) {
	if g0.ch.DidClose() {
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
	if g0.parent != nil {
		g0.parent.ConsumeError(goError)
	}

	// send the error to the channel of this stand-alone G1Group
	g0.ch.Send(goError)
}

func (g0 *GoGroup) Ch() (ch <-chan parl.GoError) { return g0.ch.Ch() }

func (g0 *GoGroup) FirstFatal() (firstFatal *parl.OnceWaiterRO) {
	g0.owLock.Lock()
	defer g0.owLock.Unlock()

	if g0.onceWaiter == nil {
		g0.onceWaiter = parl.NewOnceWaiter(context.Background())
		if g0.hadFatal.IsTrue() {
			g0.onceWaiter.Cancel()
		}
	}
	return parl.NewOnceWaiterRO(g0.onceWaiter)
}

func (g0 *GoGroup) EnableTermination(allowTermination bool) {
	if g0.isDebug.IsTrue() {
		parl.Log("goGroup%s#:EnableTermination:%t", g0.G0ID(), allowTermination)
	}
	if g0.isWaitGroupDone.IsTrue() {
		return // GoGroup is already shutdown return
	} else if allowTermination {
		if g0.noTermination.Clear() {
			// termination now allowed, it was previously blocked: remove fake Go
			g0.GoDone(g0.fakeGo.Get(), nil)
			g0.fakeGo.Put(nil)
		}
	} else if g0.noTermination.Set() {
		// termination now prevented, it was previously allowed
		// to prevent termination, add a fake thread
		g0.fakeGo.Put(g0.Go().(*Go))
	}
}

func (g0 *GoGroup) IsEnableTermination() (mayTerminate bool) {
	mayTerminate = !g0.noTermination.IsTrue()
	return
}

func (g0 *GoGroup) Threads() (threads []parl.ThreadData) {
	// the pointer can be updated at any time, but the value does not change
	list := g0.gos.List()
	threads = make([]parl.ThreadData, len(list))
	for i, tp := range list {
		threads[i] = tp
	}
	return
}

func (g0 *GoGroup) NamedThreads() (threads []parl.ThreadData) {
	// the pointer can be updated at any time, but the value does not change
	list := g0.gos.List()

	// remove unnamed threads
	for i := 0; i < len(list); {
		if list[i].label == "" {
			list = slices.Delete(list, i, i+1)
		} else {
			i++
		}
	}

	// sort pointers
	slices.SortFunc(list, g0.cmpNames)

	// return slice of values
	threads = make([]parl.ThreadData, len(list))
	for i, tp := range list {
		threads[i] = tp
	}
	return
}

func (g0 *GoGroup) SetDebug(debug bool) {
	if debug {
		g0.isDebug.Set()
	} else {
		g0.isDebug.Clear()
	}
}

func (g0 *GoGroup) cmpNames(a *ThreadData, b *ThreadData) (result bool) {
	return a.label < b.label
}

func (g0 *GoGroup) setFirstFatal() {
	g0.owLock.Lock()
	defer g0.owLock.Unlock()

	if g0.onceWaiter == nil {
		return // FirstFatal not invoked return
	}

	g0.onceWaiter.Cancel()
}

// isEnd determines if this goGroup has ended
//   - if goGroup has error channel, the goGroup ends when its error channel closes
//   - — this is true for goGroups without a parent
//   - — this is true for sub-goGroups only collecting fatal errors
//   - — for subGo, it ends when all threads have exited
//   - if goGroup otherwise has
func (g0 *GoGroup) isEnd() (isEnd bool) {

	// SubGo termination flag
	if !g0.hasErrorChannel.IsTrue() {
		return g0.isWaitGroupDone.IsTrue()
	}

	// others is by error channel — wait until all errors have been read
	return g0.ch.IsClosed()
}

func (g0 *GoGroup) listThreads() (threads []*ThreadData) {
	return g0.gos.List()
}

// "goGroup#1" "subGroup#2" "subGo#3"
func (g0 *GoGroup) typeString() (s string) {
	if g0.parent == nil {
		s = "goGroup"
	} else if g0.isSubGroup.IsTrue() {
		s = "subGroup"
	} else {
		s = "subGo"
	}
	return s + "#" + g0.goEntityID.G0ID().String()
}

// "goGroup#1:2" "subGroup#2:1" "subGo#3:0"
func (g0 *GoGroup) goGroupState() (s string) {
	adds, dones := g0.goWaitGroup.wg.Counters()
	return g0.typeString() + ":" + strconv.Itoa(adds-dones)
}

// g1Group#3threads:1(1)g0.TestNewG1Group-g1-group_test.go:60
func (g0 *GoGroup) String() (s string) {
	return g0.typeString() + parl.Sprintf("_threads:%s_New:%s",
		g0.goWaitGroup.wg.String(),
		g0.goEntityID.creator.Short(),
	)
}
