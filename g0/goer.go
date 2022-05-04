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
	// grCheckThreadFrames is how many gframes to skip: Goer.checkState + Goer.add/done/AddError/Go
	grCheckThreadFrames = 2
)

type Goer struct {
	GoerGroup      // Ch() IsExit() Wait() Cancel() Context()
	conduit        parl.ErrorConduit
	exitAction     parl.ExitAction
	gc             *GoGroup
	lock           sync.Mutex
	parentID       parl.ThreadID              // behind lock
	threadID       parl.ThreadID              // behind lock
	otherIDs       map[parl.ThreadID]struct{} // behind lock
	addLocation    pruntime.CodeLocation      // behind lock
	createLocation pruntime.CodeLocation      // behind lock
	funcLocation   pruntime.CodeLocation      // behind lock
}

func NewGoer(
	conduit parl.ErrorConduit,
	exitAction parl.ExitAction,
	index parl.GoIndex,
	gc *GoGroup,
	parentID parl.ThreadID,
	addLocation *pruntime.CodeLocation) (goer parl.Goer) {

	return &Goer{
		GoerGroup: GoerGroup{
			waiterr:          waiterr{wg: &parl.WaitGroup{}, index: index},
			cancelAndContext: *newCancelAndContext(gc.Context()),
		},
		conduit:     conduit,
		exitAction:  exitAction,
		gc:          gc,
		parentID:    parentID,
		otherIDs:    map[parl.ThreadID]struct{}{},
		addLocation: *addLocation,
	}
}

func (gr *Goer) Go() (g0 parl.Go) {
	gr.checkState("Go")
	gr.add(1)
	return NewGo(
		gr.AddError,
		gr.add,
		gr.done,
		gr.Context,
		gr.Cancel,
	)
}

func (gr *Goer) AddError(err error) {
	gr.checkState("Go.AddError")

	// package error
	if err == nil {
		return
	}
	goError := NewGoError(err, parl.GeNonFatal, gr)

	// send GoError on shared channel
	if gr.conduit == parl.EcSharedChan {
		gr.gc.send(goError)
		return
	}

	// send err on dedicated error channel
	gr.send(goError)
}

func (gr *Goer) add(delta int) {
	gr.checkState("Go.AddError")
	gr.waiterr.add(delta)
}

func (gr *Goer) done(errp *error) {
	parl.Debug("Goer.done" + gr.string(errp))
	gr.checkState("Go.Done")

	// send error
	isDone, goError := gr.doneAndErrp(errp)
	err := goError.GetError()
	if gr.conduit == parl.EcSharedChan {
		gr.gc.send(goError)
	} else {

		// send ThreadResult on dedicated error channel
		gr.send(goError)
	}

	if !isDone {
		return // it is a sub-thread exit
	}

	// mark Done
	gr.close()
	gr.gc.exitAction(err, gr.exitAction, gr.index)
}

func (gr *Goer) checkState(action string) (didClose bool) {
	if didClose = gr.didClose(); didClose {
		panic(perrors.Errorf(action+" after close\nSTACK: %s\n", pruntime.DebugStack(0)))
	}

	stack := goid.NewStack(grCheckThreadFrames)
	gr.lock.Lock()
	defer gr.lock.Unlock()

	// ensure we have a threadID
	if gr.threadID == "" {
		gr.threadID = stack.ID()
		gr.createLocation = *stack.Creator()
		gr.funcLocation = *stack.Frames()[len(stack.Frames())-1].Loc()
		return
	}

	// collect additional thread IDs
	if stack.ID() == gr.threadID {
		return
	}
	if _, ok := gr.otherIDs[stack.ID()]; !ok {
		gr.otherIDs[stack.ID()] = struct{}{}
	}

	return
}

/*
String outputs:
#[index] [waitcount]([adds]):
—lines: #[Goer-index] [waitcount]([adds]) ID: [thread-ID or parent-thread-ID] [create location].

index is a main-invocation-unique numeric identifier for the Goer.
waitcount is the oustanding number of done invocations for the Goer.
adds is the number of adds invoked for the Goer.
*/
func (gr *Goer) String() (s string) {
	sList := []string{fmt.Sprintf("#%d", gr.index)}
	adds, dones := gr.counters()
	sList = append(sList, fmt.Sprintf("%d(%d)", adds-dones, adds))
	if gr.threadID != "" {
		sList = append(sList, "ID: "+gr.threadID.String())
	} else {
		sList = append(sList, "pID: "+gr.parentID.String())
	}
	if gr.funcLocation.FuncName != "" {
		sList = append(sList, gr.funcLocation.Short())
	} else {
		sList = append(sList, "add: "+gr.addLocation.Short())
	}

	return strings.Join(sList, "\x20")
}
