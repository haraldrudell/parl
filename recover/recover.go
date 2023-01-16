/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package recover

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	recAnnStackFrames    = 1
	recRecStackFrames    = 2
	recEnsureErrorFrames = 2
	// Recover() and Recover2() are deferred functions invoked on panic
	// Because the functions are directly invoked by runtime panic code,
	// there are no intermediate stack frames between Recover*() and runtime.panic*.
	// therefore, the Recover stack frame must be included in the error stack frames
	// recover2() + processRecover() + ensureError() == 3
	recProcessRecoverFrames = 3
)

// Recover recovers from a panic invoking a function no more than once.
// If there is *errp does not hold an error and there is no panic, onError is not invoked.
// Otherwise, onError is invoked exactly once.
// *errp is updated with a possible panic.
func Recover(annotation string, errp *error, onError func(error)) {
	recover2(annotation, errp, onError, false, recover())
}

// Recover2 recovers from a panic and may invoke onError multiple times.
// onError is invoked if there is an error at *errp and on a possible panic.
// *errp is updated with a possible panic.
func Recover2(annotation string, errp *error, onError func(error)) {
	recover2(annotation, errp, onError, true, recover())
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

// AddToPanic ensures that a recover() value is an error or nil.
func EnsureError(panicValue interface{}) (err error) {
	return ensureError(panicValue, recEnsureErrorFrames)
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

// AddToPanic takes a recover() value and adds it to additionalErr.
func AddToPanic(panicValue interface{}, additionalErr error) (err error) {
	if err = EnsureError(panicValue); err == nil {
		return additionalErr
	}
	if additionalErr == nil {
		return
	}
	return perrors.AppendError(err, additionalErr)
}

// HandlePanic recovers from panic in fn returning error.
func HandlePanic(fn func()) (err error) {
	defer Recover(Annotation(), &err, nil)

	fn()
	return
}

// HandleErrp recovers from a panic in fn storing at *errp.
// HandleErrp is deferable.
func HandleErrp(fn func(), errp *error) {
	defer Recover(Annotation(), errp, nil)

	fn()
}

// HandleParlError recovers from panic in fn invoking an error callback.
// HandleParlError is deferable
// storeError can be the thread-safe perrors.ParlError.AddErrorProc()
func HandleParlError(fn func(), storeError func(err error)) {
	defer Recover(Annotation(), nil, storeError)

	fn()
}

// perrors.ParlError.AddErrorProc can be used with HandleParlError
var _ func(err error) = (&perrors.ParlError{}).AddErrorProc
