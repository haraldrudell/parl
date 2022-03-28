/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"

	"github.com/haraldrudell/parl/error116"
	"github.com/haraldrudell/parl/runt"
)

const (
	recAnnStackFrames = 1
	recRecStackFrames = 2
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

func recover2(annotation string, errp *error, onError func(error), multiple bool, v interface{}) {

	// ensure non-empty annotation
	if annotation == "" {
		annotation = runt.NewCodeLocation(recRecStackFrames).PackFunc() + ": panic:"
	}

	// consume *errp
	var err error
	if errp != nil {
		if err = *errp; err != nil && multiple {
			invokeOnError(onError, err)
		}
	}

	// consume recover()
	if e := processRecover(annotation, v); e != nil {
		if multiple {
			invokeOnError(onError, e)
		} else {
			err = error116.AppendError(err, e)
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

func Annotation() (annotation string) {
	return fmt.Sprintf("Recover from panic in %s:", runt.NewCodeLocation(recAnnStackFrames).PackFunc())
}

// processRecover ensures non-nil result to be error with Stack
func processRecover(annotation string, panicValue interface{}) (err error) {
	if err = EnsureError(panicValue); err == nil {
		return
	}

	// annotate
	if annotation != "" {
		err = Errorf("%s '%w'", annotation, err)
	}
	return
}

func EnsureError(panicValue interface{}) (err error) {
	if panicValue == nil {
		return
	}

	// ensure value to be error
	var ok bool
	if err, ok = panicValue.(error); !ok {
		err = Errorf("non-error value: %T %+[1]v", panicValue)
	}
	return
}

func AddToPanic(panicValue interface{}, additionalErr error) (err error) {
	if err = EnsureError(panicValue); err == nil {
		return additionalErr
	}
	if additionalErr == nil {
		return
	}
	return error116.AppendError(err, additionalErr)
}

// HandlePanic recovers from panics when executing fn.
// A panic is returned in err
func HandlePanic(fn func()) (err error) {
	defer Recover(Annotation(), &err, nil)
	fn()
	return
}

// HandleErrp recovers from panics when executing fn.
// A panic is stored at errp using error116.AppendError()
func HandleErrp(fn func(), errp *error) {
	defer Recover(Annotation(), errp, nil)
	fn()
}

// HandleErrp recovers from panics when executing fn.
// A panic is provided to the storeError function.
// storeError can be the thread-safe error116.ParlError.AddErrorProc()
func HandleParlError(fn func(), storeError func(error)) {
	defer Recover(Annotation(), nil, storeError)
	fn()
}

var _ = (&error116.ParlError{}).AddErrorProc
