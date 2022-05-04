/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"time"
)

// GoGroup manages any number of go statements.
// Exit action and error routing is configurable per go statement.
// Each go statement can send errors, receive cancel, initiate cancel and be waited on.
// The GoGroup and each go statement are cancelable.
// The GoGroup and each go statement can be waited on.
type GoGroup interface {
	// Add indicates a new goroutine about to be launched.
	// conduit indicates how errors will be propagated from
	// the goroutine.
	// exitAction describes what actions the GoCreator object
	// will take upon goroutine exit
	Add(conduit ErrorConduit, exitAction ExitAction) (goer Goer)
	// WaitPeriod waits for all goroutines managed by this GoCreator
	// to exit, periodically printing a description of the goroutines
	// that have yet to exit
	WaitPeriod(duration ...time.Duration)
	GoManager
}

// Goer manages one or more go statements.
// Goer is obtained from a GoGroup to manage a single go statement or
// from a GoerFactory to manage any number of go statements uniformly or
// from a GoError.
// The managed go statements can send errors, receive cancel, initiate cancel and be waited on.
// Cancel and Wait are separate for the Goer and only pertains to the managed go statements.
// The wait mechanic used is observable.
type Goer interface {
	// Go returns a Go object to be provided to a go statement.
	// A Goer obained from GoGroup is only intended to manage one go statement.
	Go() (g0 Go)
	// AddError emits a GeNonFatal error on the error channel
	AddError(err error)
	GoManager
}

type GoManager interface {
	// Ch is a channel that will close once all go statements have exited.
	// If the Goer was obtained from a GoGroup, Ch only emits data for EcErrChan.
	// Ch emits errors and exit results as they occur.
	Ch() (ch <-chan GoError)
	// IsExit indicates whether all go statements and Go.Add incovations
	// has exited.
	IsExit() (isExit bool)
	// Wait waits for all go statements and Go.Add invocations
	Wait()
	// Cancel indicates to all threads managed by this Goer that
	// work done on behalf of this context should be canceled
	Cancel()
	// Context will cancel when work done on behalf of this context
	// should be canceled
	Context() (ctx context.Context)
	String() (s string)
}

// Go offers all-in-one functions for a single go statement initiating goroutine execution.
//  AddError sends non-fatal errors ocurring during goroutine execution.
//  Done allows to provide outcome and for callers to wait for the goroutine.
//  Add allows for additional sub-threads
//  Context provide a Done channel and Err to determine if the goroutine should cancel.
//  SubGo allows for sub-threads with separate cancelation and wait
type Go interface {
	// Register performs no function but allows the Go object to collect
	// information on the new thread
	Register()
	// Add allows for a goroutine to have the caller wait for
	// additional internal goroutines.
	Add(delta int)
	// AddError allows a goroutine to send non-fatal errors
	AddError(err error)
	// Done indicates that a goroutine has finished.
	// err nil typically means successful exit.
	// Done is deferrable.
	// If the waitGroup is not done, a GePreDoneExit status is sent.
	// If Done is for the final goroutine, GeExit is sent.
	Done(errp *error)
	// Cancel allows for the goroutine or its sub-threads to initiate local cancel
	Cancel()
	// Context will cancel when work done on behalf of this context
	// should be canceled
	Context() (ctx context.Context)
	// SubGo allows a sub-group of threads that can be canceled and waited for separately.
	// Subgo has access to the AddError error sink.
	// Subgo has its own sub-context and waitgroup.
	// Subgo Add and Done are duplicated.
	SubGo() (subGo SubGo)
}

// SubGo allows an executing go statement provided a Go object to have sub-thread go statements.
// SubGo is a Go with individual cancel and obervable TraceGroup.
// Wait waits for all sub-threads to exit.
// Cancel affects all sub-threads.
type SubGo interface {
	// Go: SubGo behaves like a Go
	Go
	// Wait allows a thread to wait for (many) sub-threads
	Wait()
	String() (s string)
}

// GoError is an error or a thread exit associated with a goroutine
// Goer returns the Goer object handling the goroutine that originated the error
type GoError interface {
	error // Error() string
	// GetError retrieves the original error value
	GetError() (err error)
	// Source describes the source and significance of the error
	Source() (source GoErrorSource)
	// Goer provides the Goer object associated with the goroutine
	// causing the error
	Goer() (goer Goer)
	String() (s string)
}

const (
	// GeNonFatal indicates a non-fatal error ocurring during processing.
	// err is non-nil
	GeNonFatal GoErrorSource = iota + 1
	// GePreDoneExit indicates an exit value of a subordinate goroutine,
	// other than the final exit of the last running goroutine.
	// err may be nil
	GePreDoneExit
	// GeExit indicates exit of the last goroutine.
	// err may be nil.
	// The error channel may close after GeExit.
	GeExit
)

const (
	// EcSharedChan emits error on a shared error channel of the GoCreator object
	EcSharedChan ErrorConduit = iota + 1
	// EcErrChan emits error on an individual channel of the Goer object
	EcErrChan
	// TODO 220418 ECErrorStore stores error in a perrors.ErrorStore of the GoCreator object
	//ECErrorStore
)

const (
	// ExCancelOnExit cancels the GoCreator context ie. all actions on behalf of the
	// GoCreator if the goroutine exits
	ExCancelOnExit ExitAction = iota + 1
	// ExIgnoreExit takes no action on goruotine exit
	ExIgnoreExit
	// ExCancelOnFailure cancels the GoCreator context ie. all actions on behalf of the
	// GoCreator if the goroutine exits with error
	ExCancelOnFailure
)

type GoIndex int
type ErrorConduit uint8
type ExitAction uint8

type GoGroupFactory interface {
	NewGoGroup(ctx context.Context) (goGroup GoGroup)
}

type GoerFactory interface {
	NewGoer(ctx context.Context) (goer Goer)
}
