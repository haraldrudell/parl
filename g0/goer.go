/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"fmt"
	"strings"
	"sync"

	"github.com/haraldrudell/parl"

	"github.com/haraldrudell/parl/goid"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	grCheckThreadFrames = 3
)

type GoerDo struct {
	conduit        parl.ErrorConduit
	index          parl.GoIndex
	exitAction     parl.ExitAction
	gc             *GoCreatorDo
	errCh          chan error
	lock           sync.Mutex
	parentID       parl.ThreadID // behind lock
	threadID       parl.ThreadID // behind lock
	otherIDs       map[parl.ThreadID]struct{}
	addLocation    pruntime.CodeLocation // behind lock
	createLocation pruntime.CodeLocation // behind lock
	funcLocation   pruntime.CodeLocation // behind lock
	gr             *GoerRuntime          // behind lock
}

func NewGoer(
	conduit parl.ErrorConduit,
	exitAction parl.ExitAction,
	index parl.GoIndex,
	gc *GoCreatorDo,
	parentID parl.ThreadID,
	addLocation *pruntime.CodeLocation) (goer parl.Goer) {
	gd := GoerDo{
		conduit:     conduit,
		exitAction:  exitAction,
		index:       index,
		gc:          gc,
		errCh:       make(chan error),
		parentID:    parentID,
		otherIDs:    map[parl.ThreadID]struct{}{},
		addLocation: *addLocation,
		gr:          &GoerRuntime{},
	}
	gd.gr.wg.Add(1)
	gd.gr.g0 = NewGo(
		gd.errorReceiver,
		gd.add,
		gd.done,
		gc.ctx,
	)
	return &gd
}

func (gr *GoerDo) Go() (g0 parl.Go) {
	goerRuntime := gr.getGoerRuntime(false)
	if goerRuntime == nil {
		panic(perrors.New("Goer.GO after thread exit"))
	}
	return goerRuntime.g0
}

func (gr *GoerDo) Chan() (ch <-chan error) {
	return gr.errCh
}

func (gr *GoerDo) Wait() {
	goerRuntime := gr.getGoerRuntime(false)
	if goerRuntime != nil {
		goerRuntime.wg.Wait()
	}
}

func (gr *GoerDo) add(delta int) {
	goerRuntime := gr.getGoerRuntime(true)
	if goerRuntime == nil {
		panic(perrors.New("Go.Add after thread exit"))
	}
	goerRuntime.wg.Add(delta)
}

func (gr *GoerDo) done(err error) {
	goerRuntime := gr.getGoerRuntime(true)
	if goerRuntime == nil {
		panic(perrors.New("Go.Done after thread exit"))
	}

	// execute Done
	goerRuntime.wg.Done()
	// isDone indicates that this thread exited
	// it ir otherwise a sub-thread exit
	isDone := goerRuntime.wg.IsZero()
	var source parl.GoErrorSource
	if isDone {
		source = parl.GeExit
	} else {
		source = parl.GePreDoneExit
	}

	// send GoError
	if gr.conduit == parl.EcSharedChan {
		gr.gc.errorReceiver(NewGoError(
			err,
			source,
			gr,
		))
	} else {

		// send ThreadResult on dedicated error channel
		gr.errCh <- parl.NewThreadResult(err)
	}

	if !isDone {
		return // it is a sub-thread exit
	}

	// mark Done
	gr.gc.exitAction(err, gr.exitAction, gr.index)

	gr.lock.Lock()
	defer gr.lock.Unlock()

	gr.gr = nil
	if parl.Closer(gr.errCh, &err); err != nil {
		gr.gc.errorReceiver(NewGoError(err, parl.GeInternal, gr))
	}
}

func (gr *GoerDo) errorReceiver(err error) {
	goerRuntime := gr.getGoerRuntime(true)
	if goerRuntime == nil {
		panic(perrors.New("Go.AddError after thread exit"))
	}

	// send GoError
	if gr.conduit == parl.EcSharedChan {
		gr.gc.errorReceiver(NewGoError(
			err,
			parl.GeNonFatal,
			gr,
		))
		return
	}

	// send err on dedicated error channel
	gr.errCh <- err
}

func (gr *GoerDo) getGoerRuntime(checkThread bool) (goerRuntime *GoerRuntime) {
	var stack *goid.Stack
	if checkThread {
		stack = goid.NewStack(grCheckThreadFrames)
	}
	gr.lock.Lock()
	defer gr.lock.Unlock()

	if checkThread {
		if gr.threadID == "" {
			gr.threadID = stack.ID
			gr.createLocation = stack.Creator
			gr.funcLocation = stack.Frames[len(stack.Frames)-1].CodeLocation
		} else if stack.ID != gr.threadID {
			if _, ok := gr.otherIDs[stack.ID]; !ok {
				gr.otherIDs[stack.ID] = struct{}{}
			}
		}
	}

	return gr.gr
}

/*
#[index] [waitcount] ([adds]) [thread-ID or parent-thread-ID] add: add-location
# go-creator index
*/
func (gr *GoerDo) String() (s string) {
	sList := []string{fmt.Sprintf("#%d", gr.index)}
	if goerRuntime := gr.getGoerRuntime(false); goerRuntime != nil {
		adds, dones := goerRuntime.wg.Counters()
		sList = append(sList, fmt.Sprintf("%d(%d)", adds-dones, adds))
	}
	if gr.threadID != "" {
		sList = append(sList, "ID: "+gr.threadID.String())
	} else {
		sList = append(sList, "pID: "+gr.parentID.String())
	}
	if gr.funcLocation.FuncName != "" {
		sList = append(sList, gr.funcLocation.PackFunc())
	} else {
		sList = append(sList, "add: "+gr.addLocation.Short())
	}

	return strings.Join(sList, "\x20")
}
