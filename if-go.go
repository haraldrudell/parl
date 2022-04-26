/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"time"
)

// GoCreator manages the life cycle of one or more goroutines
type GoCreator interface {
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
// Goer returns the Goer object handling the goroutine that sent the error
type GoError interface {
	error // Error() string
	// GetError retrieves the original error value
	GetError() (err error)
	// Source describes the source and significance of the error
	Source() (source GoErrorSource)
	// Goer provides the Goer object associated with the goroutine
	// causing the error
	Goer() (goer Goer)
}

const (
	// GeNonFatal indictaes a non-fatal error ocurring during processing
	GeNonFatal = iota + 1
	// GePreDoneExit indicates an exit value of a subordinate goroutine,
	// other than the main goroutine associated with the GoCreator object
	// err may be nil
	GePreDoneExit
	// GeExit indicates exit of the main goroutine
	// err may be nil
	GeExit
	// GeInternal indictaes an error occuring inside the Goer object
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

// Goer is the managing object for a goroutine
type Goer interface {
	// Go gets the Go object, that is handed to its goroutine on launch.
	Go() (g0 Go)
	// Chan is a channel that will close on thread exit.
	// Chan will emit errors as they occur if the thread was launched with ECErrChan
	// Chan will emit an Exit result if the thread was launched with ECErrChan
	Chan() (ch <-chan error)
	// Wait allows to wait for this exact goroutine
	Wait()
}

// parl.Go is a value provided to a goroutine allowing it
// to provide that it has finished and its result.
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
	Done(err error)
	// Context will cancel when work done on behalf of this context
	// should be canceled
	Context() (ctx context.Context)
	// SubGo allows a sub-group of threads to be cancelled and waited for separately.
	// Subgo still has access to AddError error sink.
	SubGo() (subGo SubGo)
}

// SubGo is a Go with its own CancelContext and WaitGroup.
// Wait is used by the thread invoking SubGo waiting for its sub-threads to exit.
// Cancel is used by any of the SubGo invoker and its sub-threads to cancel the group.
type SubGo interface {
	// Go: SubGo behaves like a Go
	Go
	// Wait allows a thread to wait for all sub-threads: many-to-many
	Wait()
	// Cancel allows for any thread to cancel all sub-threads: many-to-many
	Cancel()
}

type GoCreatorFactory interface {
	NewGoCreator(ctx context.Context) (goCreator GoCreator)
}
