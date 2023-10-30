/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package test

import (
	"errors"
	"fmt"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog"
	"github.com/haraldrudell/parl/pruntime"
)

// this code comes from parl

const (
	recAnnStackFrames = 1
	recRecStackFrames = 2
	// Recover() and Recover2() are deferred functions invoked on panic
	// Because the functions are directly invoked by runtime panic code,
	// there are no intermediate stack frames between Recover*() and runtime.panic*.
	// therefore, the Recover stack frame must be included in the error stack frames
	// recover2() + processRecover() + ensureError() == 3
	recProcessRecoverFrames = 3
	// extra stack frame for the parl indirection, used by 6 functions:
	// Debug() GetDebug() D() GetD() IsThisDebug() IsThisDebugN()
	logStackFramesToSkip = 1
)

var stderrLogger = plog.NewLogFrames(nil, logStackFramesToSkip)

// ErrErrpNil indicates that a function with an error pointer argument
// received an errp nil value.
//
//	if errors.Is(err, ErrErrpNil) …
var ErrErrpNil = errors.New("errp cannot be nil")

// RecoverInvocationPanic is intended to wrap callback invocations in the callee in order to
// recover from panics in the callback function.
// when an error occurs, perrors.AppendError appends the callback error to *errp.
// if fn is nil, a recovered panic results.
// if errp is nil, a panic is thrown, can be check with:
//
//	if errors.Is(err, ErrErrpNil) …
func RecoverInvocationPanic(fn func(), errp *error) {
	if errp == nil {
		panic(perrors.ErrorfPF("%w", ErrErrpNil))
	}
	defer Recover("", errp, NoOnError)

	fn()
}

// Recover recovers from a panic invoking a function no more than once.
// If there is *errp does not hold an error and there is no panic, onError is not invoked.
// Otherwise, onError is invoked exactly once.
// *errp is updated with a possible panic.
func Recover(annotation string, errp *error, onError func(error)) {
	recover2(annotation, errp, onError, false, recover())
}

// NoOnError is used with Recover to silence the default error logging
func NoOnError(err error) {}

// Annotation provides a default annotation [base package].[function]: "mypackage.MyFunc"
func Annotation() (annotation string) {
	return fmt.Sprintf("Recover from panic in %s:", pruntime.NewCodeLocation(recAnnStackFrames).PackFunc())
}

// processRecover ensures non-nil result to be error with Stack
func processRecover(annotation string, panicValue interface{}) (err error) {
	if err = ensureError(panicValue, recProcessRecoverFrames); err == nil {
		return // panicValue nil return: no error
	}

	// annotate
	if annotation != "" {
		err = perrors.Errorf("%s \x27%w\x27", annotation, err)
	}
	return
}

func ensureError(panicValue interface{}, frames int) (err error) {

	if panicValue == nil {
		return // no panic return
	}

	// ensure value to be error
	var ok bool
	if err, ok = panicValue.(error); !ok {
		err = fmt.Errorf("non-error value: %T %+[1]v", panicValue)
	}

	// ensure stack trace
	if !perrors.HasStack(err) {
		err = perrors.Stackn(err, frames)
	}

	return
}

func recover2(annotation string, errp *error, onError func(error), multiple bool, recoverValue interface{}) {
	// ensure non-empty annotation
	if annotation == "" {
		annotation = pruntime.NewCodeLocation(recRecStackFrames).PackFunc() + ": panic:"
	}

	// consume *errp
	var err error
	if errp != nil {
		if err = *errp; err != nil && multiple {
			invokeOnError(onError, err)
		}
	}

	// consume recover()
	if e := processRecover(annotation, recoverValue); e != nil {
		if multiple {
			invokeOnError(onError, e)
		} else {
			err = perrors.AppendError(err, e)
		}
	}

	// write back to *errp, do non-multiple invocation
	if err != nil {
		if errp != nil && *errp != err {
			*errp = err
		}
		if !multiple {
			invokeOnError(onError, err)
		}
	}
}

func invokeOnError(onError func(error), err error) {
	if onError != nil {
		onError(err)
	} else {
		Log("Recover: %+v\n", err)
	}
}

// Log invocations always print and output to stderr.
// if debug is enabled, code location is appended.
func Log(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}
