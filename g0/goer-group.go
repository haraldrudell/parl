/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type GoerGroup struct {
	waiterr          // Ch() IsExit() Wait() String()
	cancelAndContext // Cancel() Context()
}

func NewGoerGroup(ctx context.Context) (goer parl.Goer) {
	return &GoerGroup{
		waiterr:          waiterr{wg: &parl.TraceGroup{}, index: Index.goIndex()},
		cancelAndContext: *newCancelAndContext(ctx),
	}
}

func (gr *GoerGroup) Go() (g0 parl.Go) {
	if gr.didClose() {
		panic(perrors.New("Go after close"))
	}
	gr.add(1)
	return NewGo(
		gr.AddError,
		gr.waiterr.add,
		gr.done,
		gr.Context,
		gr.Cancel,
	)
}

func (gr *GoerGroup) AddError(err error) {
	if gr.didClose() {
		panic(perrors.New("Go.AddError after close"))
	}
	if err == nil {
		return
	}
	gr.send(NewGoError(err, parl.GeNonFatal, gr))
}

func (gr *GoerGroup) done(errp *error) {
	parl.Debug("GoerGroup.done" + gr.string(errp))
	if gr.didClose() {
		panic(perrors.New("Go.Done after close"))
	}

	isDone, goError := gr.doneAndErrp(errp)
	gr.send(goError)

	// possibly close error channel
	if isDone {
		gr.close()
	}
}

func (gr *GoerGroup) doneAndErrp(errp *error) (isDone bool, goError parl.GoError) {
	// execute done
	isDone = gr.doneBool()

	// get error value
	var err error
	if errp != nil {
		err = *errp
	} else {
		err = perrors.New("g0.Done with errp nil")
	}

	// send result
	var source parl.GoErrorSource
	if isDone {
		source = parl.GeExit
	} else {
		source = parl.GePreDoneExit
	}
	goError = NewGoError(err, source, gr)

	return
}
