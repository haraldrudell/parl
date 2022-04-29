/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"time"
)

// GoGroup manages many go statements
type GoGroup interface {
	// Add indicates a new goroutine about to be launched.
	// conduit indicates how errors will be propagated from
	// the goroutine.
	// exitAction describes what actions the GoCreator object
	// will take upon goroutine exit
	Add(conduit ErrorConduit, exitAction ExitAction) (goer Goer)
	// Warnings provides an error channel shared by goroutines that
	// do not have an individual channel
	Warnings() (ch <-chan GoError)
	// Wait waits for all goroutines managed by this GoCreator
	// to exit
	Wait()
	// WaitPeriod waits for all goroutines managed by this GoCreator
	// to exit, periodically printing a description of the goroutines
	// that have yet to exit
	WaitPeriod(duration ...time.Duration)
	IsExit() (isExit bool)
	List() (s string)
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
	// GeNonFatal indicates a non-fatal error ocurring during processing
	GeNonFatal GoErrorSource = iota + 1
	// GePreDoneExit indicates an exit value of a subordinate goroutine,
	// other than the main goroutine associated with the GoCreator object
	// err may be nil
	GePreDoneExit
	// GeExit indicates exit of the main goroutine
	// err may be nil
	GeExit
	// GeInternal indicates an error occuring inside the Goer object
	GeInternal
)

type GoErrorSource uint8

type GoIndex int

const (
	// EcSharedChan emits error on a shared error channel of the GoCreator object
	EcSharedChan ErrorConduit = iota + 1
	// EcErrChan emits error on an individual channel of the Goer object
	EcErrChan
	// TODO 220418 ECErrorStore stores error in a perrors.ErrorStore of the GoCreator object
	//ECErrorStore
)

type ErrorConduit uint8

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

type ExitAction uint8

// Goer manages a single go statement
type Goer interface {
	// Go gets the Go object, that is provided to its goroutine in a go statement
	Go() (g0 Go)
	// Ch is a channel that will close on thread exit.
	// Ch will emit errors as they occur if the thread was launched with ECErrChan
	// Ch emitting
	// Ch will emit an Exit result if the thread was launched with ECErrChan
	Ch() (ch <-chan GoError)
	// Context will cancel when work done on behalf of this context
	// should be canceled
	Context() (ctx context.Context)
	Cancel()
	// Wait allows to wait for this exact goroutine
	Wait()
	String() (s string)
}

// SubGoer manages a group of goroutines
type SubGoer interface {
	// Go gets a Go object, that is used in a go statement.
	Go() (g0 Go)
	// Ch is a channel that sends errors as they occur amd will close after GeExit
	// Ch send GeNonFatal for non-fatal errors, err is non-nil
	// Ch sends GePreDoneExit for sub-thread exits, err may be nil
	// Ch sends a final GeExit for the final exit, err may be nil
	Ch() (ch <-chan GoError)
	// Cancel cancels all goroutines managed by the Subgoer
	// AddError allows a goroutine to send non-fatal errors
	AddError(err error)
	Cancel()
	// Wait allows to wait for all goroutines to exit
	// Context will cancel when work done on behalf of this context
	// should be canceled
	Context() (ctx context.Context)
	Wait()
	IsExit() (isExit bool)
	String() (s string)
}

// parl.Go allows a goroutine to implement cancel, submit errors, communicate exit,
// be waited upon and to have sub-threads.
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
	Done(errp *error)
	// Context will cancel when work done on behalf of this context
	// should be canceled
	Context() (ctx context.Context)
	// SubGo allows a sub-group of threads to be cancelled and waited for separately.
	// Subgo still has access to AddError error sink.
	SubGo() (subGo SubGo)
}

// SubGo allows a goroutine to have sub-threads.
// SubGo is a Go with its own CancelContext and TraceGroup.
// Wait is used by the thread invoking SubGo waiting for its sub-threads to exit.
// Cancel is used by any of the SubGo invoker and its sub-threads to cancel the group.
type SubGo interface {
	// Go: SubGo behaves like a Go
	Go
	// Wait allows a thread to wait for (many) sub-threads
	Wait()
	// Cancel allows for any thread to cancel all sub-threads
	Cancel()
	String() (s string)
}

type GoCreatorFactory interface {
	NewGoCreator(ctx context.Context) (goGroup GoGroup)
}

type GoerFactory interface {
	NewGoer(ctx context.Context) (goer SubGoer)
}
