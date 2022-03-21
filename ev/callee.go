/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ev

import (
	"context"
	"errors"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/error116"
)

// CalleeContext provides for a goroutine to communicate with the caller
type CalleeContext struct {
	context.Context
	gID     GoID
	name    string
	eventTx chan<- Event
}

var _ Callee = &CalleeContext{}

// NewCallee provides a Calle implementation, data used by a managed goroutine
func NewCallee(name string, gID GoID, eventTx chan<- Event, ctx0 context.Context) (ctx Callee) {
	ctx = &CalleeContext{Context: ctx0, gID: gID, name: name, eventTx: eventTx}
	return
}

func (ctx *CalleeContext) Success() {
	ctx.eventTx <- NewExitEvent(nil, ctx.gID)
}

func (ctx *CalleeContext) Failure(err error) {
	if err == nil {
		err = error116.New("Failure without errors")
	}
	evt := NewExitEvent(err, ctx.gID)
	exitEvent := evt.(*ExitEvent)
	if errors.Is(err, context.Canceled) {
		exitEvent.IsCancel = true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		exitEvent.IsTimeout = true
	}
	ctx.eventTx <- evt
}

func (ctx *CalleeContext) Result(errp *error) {
	ctx.ResultV(errp, recover())
}

func (ctx *CalleeContext) ResultV(errp *error, recoverValue interface{}) {
	var err error

	// collect errp
	if errp != nil {
		err = *errp
	}

	// collect recoverValue, update *errp
	if recoverValue != nil {
		e := parl.EnsureError(recoverValue)
		parl.Errorf("Recover in Result for goroutine: %s: %v", ctx.name, recoverValue)
		err = error116.AppendError(err, e)
		if errp != nil && err != *errp {
			*errp = err
		}
	}

	// forward result
	if err == nil {
		ctx.Success()
	} else {
		ctx.Failure(err)
	}
}

func (ctx *CalleeContext) Thread() (name string, gID GoID) {
	name = ctx.name
	gID = ctx.gID
	return
}

// Send allows a groutine to send any event
func (ctx *CalleeContext) Send(payload interface{}) {
	ctx.eventTx <- NewDataEvent(ctx.gID, payload)
}
