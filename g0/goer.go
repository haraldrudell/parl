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
	exitAction     parl.ExitAction
	index          parl.GoIndex
	gc             *GoCreatorDo
	errCh          parl.NBChan[parl.GoError]
	lock           sync.Mutex
	parentID       parl.ThreadID              // behind lock
	threadID       parl.ThreadID              // behind lock
	otherIDs       map[parl.ThreadID]struct{} // behind lock
	addLocation    pruntime.CodeLocation      // behind lock
	createLocation pruntime.CodeLocation      // behind lock
	funcLocation   pruntime.CodeLocation      // behind lock
	gr             *GoerRuntime               // behind lock
	ctx            parl.CancelContext
}

func NewGoer(
	conduit parl.ErrorConduit,
	exitAction parl.ExitAction,
	index parl.GoIndex,
	gc *GoCreatorDo,
	parentID parl.ThreadID,
	addLocation *pruntime.CodeLocation) (goer parl.Goer) {
	goer0 := GoerDo{
		conduit:     conduit,
		exitAction:  exitAction,
		index:       index,
		gc:          gc,
		parentID:    parentID,
		otherIDs:    map[parl.ThreadID]struct{}{},
		addLocation: *addLocation,
		ctx:         parl.NewCancelContext(gc.ctx),
	}
	gruntime := GoerRuntime{
		g0: NewGo(
			goer0.errorReceiver,
			goer0.add,
			goer0.done,
			goer0.Context,
		),
	}
	gruntime.wg.Add(1)
	goer0.gr = &gruntime
	return &goer0
}

func (gr *GoerDo) Go() (g0 parl.Go) {
	goerRuntime := gr.getGoerRuntime(false)
	if goerRuntime == nil {
		panic(perrors.New("Goer.GO after thread exit"))
	}
	return goerRuntime.g0
}

func (gr *GoerDo) Ch() (ch <-chan parl.GoError) {
	return gr.errCh.Ch()
}

func (gr *GoerDo) Context() (ctx context.Context) {
	return gr.ctx
}

func (gr *GoerDo) Cancel() {
	gr.ctx.Cancel()
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
	goError := NewGoError(
		err,
		source,
		gr,
	)
	if gr.conduit == parl.EcSharedChan {
		gr.gc.errorReceiver(goError)
	} else {

		// send ThreadResult on dedicated error channel
		gr.errCh.Send(goError)
	}

	if !isDone {
		return // it is a sub-thread exit
	}

	// mark Done
	gr.gc.exitAction(err, gr.exitAction, gr.index)

	gr.lock.Lock()
	defer gr.lock.Unlock()

	gr.gr = nil
	gr.errCh.Close()
}

func (gr *GoerDo) errorReceiver(err error) {
	goerRuntime := gr.getGoerRuntime(true)
	if goerRuntime == nil {
		panic(perrors.New("Go.AddError after thread exit"))
	}

	// send GoError
	goError := NewGoError(
		err,
		parl.GeNonFatal,
		gr,
	)
	if gr.conduit == parl.EcSharedChan {
		gr.gc.errorReceiver(goError)
		return
	}

	// send err on dedicated error channel
	gr.errCh.Send(goError)
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
