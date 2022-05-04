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
	GoerGroup
	conduit          parl.ErrorConduit
	exitAction       parl.ExitAction
	index            parl.GoIndex
	gc               *GoGroup
	waiterr          // Ch() IsExit() Wait()
	lock             sync.Mutex
	parentID         parl.ThreadID              // behind lock
	threadID         parl.ThreadID              // behind lock
	otherIDs         map[parl.ThreadID]struct{} // behind lock
	addLocation      pruntime.CodeLocation      // behind lock
	createLocation   pruntime.CodeLocation      // behind lock
	funcLocation     pruntime.CodeLocation      // behind lock
	cancelAndContext                            // Cancel() Context()
}

func NewGoer(
	conduit parl.ErrorConduit,
	exitAction parl.ExitAction,
	index parl.GoIndex,
	gc *GoGroup,
	parentID parl.ThreadID,
	addLocation *pruntime.CodeLocation) (goer parl.Goer) {

	return &GoerDo{
		GoerGroup: GoerGroup{
			waiterr:          waiterr{wg: &parl.WaitGroup{}},
			cancelAndContext: *newCancelAndContext(gc.Context()),
		},
		conduit:     conduit,
		exitAction:  exitAction,
		index:       index,
		gc:          gc,
		parentID:    parentID,
		otherIDs:    map[parl.ThreadID]struct{}{},
		addLocation: *addLocation,
	}
}

func (gr *GoerDo) AddError(err error) {
	gr.getGoerRuntime("Go.AddError")

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

func (gr *GoerDo) add(delta int) {
	gr.getGoerRuntime("Go.AddError")
	gr.waiterr.add(delta)
}

func (gr *GoerDo) done(errp *error) {
	gr.getGoerRuntime("Go.Done")

	// execute Done
	isDone := gr.doneBool()

	// get error value
	var err error
	if errp != nil {
		err = *errp
	} else {
		err = perrors.New("g0.Done with errp nil")
	}

	// isDone indicates that this thread exited
	// it ir otherwise a sub-thread exit
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

func (gr *GoerDo) getGoerRuntime(action string) (didClose bool) {
	didClose = gr.didClose()
	defer func() {
		if didClose {
			panic(perrors.New(action + " after close"))
		}
	}()

	stack := goid.NewStack(grCheckThreadFrames)
	gr.lock.Lock()
	defer gr.lock.Unlock()

	// unitialize thread data
	if gr.threadID == "" {
		gr.threadID = stack.ID()
		gr.createLocation = *stack.Creator()
		gr.funcLocation = *stack.Frames()[len(stack.Frames())-1].Loc()
		return
	}

	// collect thread IDs
	if stack.ID() == gr.threadID {
		return
	}
	if _, ok := gr.otherIDs[stack.ID()]; !ok {
		gr.otherIDs[stack.ID()] = struct{}{}
	}

	return
}

/*
#[index] [waitcount] ([adds]) [thread-ID or parent-thread-ID] add: add-location
# go-creator index
*/
func (gr *GoerDo) String() (s string) {
	sList := []string{fmt.Sprintf("#%d", gr.index)}
	adds, dones := gr.counters()
	sList = append(sList, fmt.Sprintf("%d(%d)", adds-dones, adds))
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
