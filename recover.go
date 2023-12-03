/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

// Recover recovers panic using deferred annotation
//   - Recover creates a single aggregate error of *errp and any panic
//   - if onError non-nil, the function is invoked zero or one time with the aggregate error
//   - if onError nil, the error is logged to standard error
//   - if errp is non-nil, it is updated with any aggregate error
//
// Usage:
//
//	func someFunc() (err error) {
//	  defer parl.Recover(func() parl.DA { return parl.A() }, &err, parl.NoOnError)
func Recover(deferredLocation func() DA, errp *error, onError OnError) {
	doRecovery(noAnnotation, deferredLocation, errp, onError, recoverOnErrrorOnce, noIsPanic, recover())
}

// Recover2 recovers panic using deferred annotation
//   - if onError non-nil, the function is invoked zero, one or two times with any error in *errp and any panic
//   - if onError nil, errors are logged to standard error
//   - if errp is non-nil:
//   - — if *errp was nil, it is updated with any panic
//   - — if *errp was non-nil, it is updated with any panic as an aggregate error
//
// Usage:
//
//	func someFunc() (err error) {
//	  defer parl.Recover2(func() parl.DA { return parl.A() }, &err, parl.NoOnError)
func Recover2(deferredLocation func() DA, errp *error, onError OnError) {
	doRecovery(noAnnotation, deferredLocation, errp, onError, recoverOnErrrorMultiple, noIsPanic, recover())
}

// RecoverAnnotation is like Recover but with fixed-string annotation
func RecoverAnnotation(annotation string, errp *error, onError OnError) {
	doRecovery(annotation, noDeferredAnnotation, errp, onError, recoverOnErrrorOnce, noIsPanic, recover())
}

// nil OnError function
//   - public for RecoverAnnotation
var NoOnError OnError

// OnError is a function that receives error values from an errp error pointer or a panic
type OnError func(err error)

const (
	// counts the frames in [parl.A]
	parlAFrames = 1
	// counts the stack-frame in [parl.processRecover]
	processRecoverFrames = 1
	// counts the stack-frame of [parl.doRecovery] and [parl.Recover] or [parl.Recover2]
	//	- but for panic detector to work, there must be one frame after
	//		runtime.gopanic, so remove one frame
	doRecoveryFrames = 2 - 1
	// fixed-string annotation is not present
	noAnnotation = ""
)

const (
	// indicates onError to be invoked once for all errors
	recoverOnErrrorOnce OnErrorStrategy = iota
	// indicates onError to be invoked once per error
	recoverOnErrrorMultiple
	// do not invoke onError
	recoverOnErrrorNone
)

// how OnError is handled: recoverOnErrrorOnce recoverOnErrrorMultiple recoverOnErrrorNone
type OnErrorStrategy uint8

// indicates deferred annotation is not present
var noDeferredAnnotation func() DA

// DA is the value returned by a deferred code location function
type DA *pruntime.CodeLocation

// contains a deferred code location for annotation
type annotationLiteral func() DA

// A is a thunk returning a deferred code location
func A() DA { return pruntime.NewCodeLocation(parlAFrames) }

// noIsPanic is a stand-in nil value when noPanic is not present
var noIsPanic *bool

// doRecovery implements recovery for Recovery andd Recovery2
func doRecovery(annotation string, deferredAnnotation annotationLiteral, errp *error, onError OnError, onErrorStrategy OnErrorStrategy, isPanic *bool, recoverValue interface{}) {
	if onErrorStrategy == recoverOnErrrorNone {
		if errp == nil {
			panic(NilError("errp"))
		}
	} else if errp == nil && onError == nil {
		panic(NilError("both errp and onError"))
	}

	// build aggregate error in err
	var err error
	if errp != nil {
		err = *errp
		// if onError is to be invoked multiple times,
		// and *errp contains an error,
		// invoke onError or Log to standard error
		if err != nil && onErrorStrategy == recoverOnErrrorMultiple {
			invokeOnError(onError, err) // invoke onError or parl.Log
		}
	}

	// consume recover()
	if recoverValue != nil {
		if isPanic != nil {
			*isPanic = true
		}
		if annotation == noAnnotation {
			annotation = getDeferredAnnotation(deferredAnnotation)
		}
		var panicError = processRecoverValue(annotation, recoverValue, doRecoveryFrames)
		err = perrors.AppendError(err, panicError)
		if onErrorStrategy == recoverOnErrrorMultiple {
			invokeOnError(onError, panicError)
		}
	}

	// if err now contains any error
	if err != nil {
		// if errp non-nil:
		//	- err was obtained from *errp
		//	- err may now be panicError or have had panicError appended
		//	- overwrite back to non-nil errp
		if errp != nil && *errp != err {
			*errp = err
		}
		// if OnError is once, invoke onError or Log with the aggregate error
		if onErrorStrategy == recoverOnErrrorOnce {
			invokeOnError(onError, err)
		}
	}
}

// getDeferredAnnotation obtains annotation from a deferred annotation function literal
func getDeferredAnnotation(deferredAnnotation annotationLiteral) (annotation string) {
	if deferredAnnotation != nil {
		if da := deferredAnnotation(); da != nil {
			var cL = (*pruntime.CodeLocation)(da)
			// single word package name
			var packageName = cL.Package()
			// recoverDaPanic.func1: hosting function name and a derived name for the function literal
			var funcName = cL.FuncIdentifier()
			// removed “.func1” suffix
			if index := strings.LastIndex(funcName, "."); index != -1 {
				funcName = funcName[:index]
			}
			annotation = fmt.Sprintf("panic detected in %s.%s:",
				packageName,
				funcName,
			)
		}
	}
	if annotation == "" {
		// default annotation cannot be obtained
		//	- the deferred Recover function is invoked directly from rutine, eg. runtime.gopanic
		//	- therefore, use fixed string
		annotation = "recover from panic:"
	}

	return
}

// invokeOnError invokes an onError function or logs to standard error if onError is nil
func invokeOnError(onError OnError, err error) {
	if onError != nil {
		onError(err)
		return
	}
	Log("Recover: %+v\n", err)
}

// processRecoverValue returns an error value with stack from annotation and panicValue
//   - annotation is non-empty annotation indicating code loction or action
//   - panicValue is non-nil value returned by built-in recover function
func processRecoverValue(annotation string, panicValue interface{}, frames int) (err error) {
	if frames < 0 {
		frames = 0
	}
	return perrors.Errorf("%s “%w”",
		annotation,
		ensureError(panicValue, frames+processRecoverFrames),
	)
}
