/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

// Recover recovers panic using deferred annotation
//   - Recover creates a single aggregate error of *errp and any panic
//   - if onError non-nil, the function is invoked zero or one time with the aggregate error
//   - if onError nil, the error is logged to standard error
//   - if errp is non-nil, it is updated with any aggregate error
//   - parl recover options:
//   - — [RecoverErr]: aggregates to error pointer with enclosing function location, optional panic flag
//   - — [Recover]: aggregates to error pointer with enclosing function location, optional single-invocation [parl.ErrorSink]
//   - — [Recover2]: aggregates to error pointer with enclosing function location, optional multiple-invocation [parl.ErrorSink]
//   - — [RecoverAnnotation]: aggregates to error pointer with fixed-string annotation, optional single-invocation [parl.ErrorSink]
//   - — [PanicToErr]: aggregates to error pointer with generic annotation, optional panic flag
//   - — preferrably: RecoverErr, Recover or Recover2 should be used to provide the package and function name
//     of the enclosing function for the defer-statement that invoked recover
//   - — PanicToErr and RecoverAnnotation cannot provide where in the stack trace recover was invoked
//
// Usage:
//
//	func someFunc() (err error) {
//	  defer parl.Recover(func() parl.DA { return parl.A() }, &err, parl.NoOnError)
func Recover(deferredLocation func() DA, errp *error, errorSink ...ErrorSink1) {
	doRecovery(noAnnotation, deferredLocation, errp, recoverOnErrrorOnce, noIsPanic, recover(), errorSink...)
}

// Recover2 recovers panic using deferred annotation
//   - if onError non-nil, the function is invoked zero, one or two times with any error in *errp and any panic
//   - if onError nil, errors are logged to standard error
//   - if errp is non-nil:
//   - — if *errp was nil, it is updated with any panic
//   - — if *errp was non-nil, it is updated with any panic as an aggregate error
//   - parl recover options:
//   - — [RecoverErr]: aggregates to error pointer with enclosing function location, optional panic flag
//   - — [Recover]: aggregates to error pointer with enclosing function location, optional single-invocation [parl.ErrorSink]
//   - — [Recover2]: aggregates to error pointer with enclosing function location, optional multiple-invocation [parl.ErrorSink]
//   - — [RecoverAnnotation]: aggregates to error pointer with fixed-string annotation, optional single-invocation [parl.ErrorSink]
//   - — [PanicToErr]: aggregates to error pointer with generic annotation, optional panic flag
//   - — preferrably: RecoverErr, Recover or Recover2 should be used to provide the package and function name
//     of the enclosing function for the defer-statement that invoked recover
//   - — PanicToErr and RecoverAnnotation cannot provide where in the stack trace recover was invoked
//
// Usage:
//
//	func someFunc() (err error) {
//	  defer parl.Recover2(func() parl.DA { return parl.A() }, &err, parl.NoOnError)
func Recover2(deferredLocation func() DA, errp *error, errorSink ...ErrorSink1) {
	doRecovery(noAnnotation, deferredLocation, errp, recoverOnErrrorMultiple, noIsPanic, recover(), errorSink...)
}

// RecoverAnnotation is like Recover but with fixed-string annotation
//   - default annotation: “recover from panic:”
//   - parl recover options:
//   - — [RecoverErr]: aggregates to error pointer with enclosing function location, optional panic flag
//   - — [Recover]: aggregates to error pointer with enclosing function location, optional single-invocation [parl.ErrorSink]
//   - — [Recover2]: aggregates to error pointer with enclosing function location, optional multiple-invocation [parl.ErrorSink]
//   - — [RecoverAnnotation]: aggregates to error pointer with fixed-string annotation, optional single-invocation [parl.ErrorSink]
//   - — [PanicToErr]: aggregates to error pointer with generic annotation, optional panic flag
//   - — preferrably: RecoverErr, Recover or Recover2 should be used to provide the package and function name
//     of the enclosing function for the defer-statement that invoked recover
//   - — PanicToErr and RecoverAnnotation cannot provide where in the stack trace recover was invoked
//
// Usage:
//
//	func someFunc() (err error) {
//	  defer parl.RecoverAnnotation("property " + property, func() parl.DA { return parl.A() }, &err, parl.NoOnError)
func RecoverAnnotation(annotation string, deferredLocation func() DA, errp *error, errorSink ...ErrorSink1) {
	doRecovery(annotation, deferredLocation, errp, recoverOnErrrorOnce, noIsPanic, recover(), errorSink...)
}

// argument to Recover: no error aggregation
var NoErrp *error

// nil OnError function
//   - public for RecoverAnnotation
//
// deprecated: use [ErrorSink]
var NoOnError OnError

// OnError is a function that receives error values from an errp error pointer or a panic
//
// deprecated: use [ErrorSink]
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

// DA is the value returned by a deferred code location function
type DA *pruntime.CodeLocation

// contains a deferred code location for annotation
type annotationLiteral func() DA

// A is a thunk returning a deferred code location
func A() DA { return pruntime.NewCodeLocation(parlAFrames) }

// noIsPanic is a stand-in nil value when noPanic is not present
var noIsPanic *bool

// doRecovery implements recovery for Recovery andd Recovery2
//   - annotation: typically empty, can be string of some distinguishing property
//   - — “copy command i/o stderr”
//   - deferredAnnotation if a function-literal thunk present in the defer statement
//   - — provides in which enclosing function the defer-statement invoking recovery is located
//   - errp: optional error aggregation, default none
//   - errorSink: optional errorSink, default errors are printed to standard error
//   - isPanic: optional panic-flag
//   - recoverValue: the value recover() returned
func doRecovery(annotation string, deferredAnnotation annotationLiteral, errp *error, onErrorStrategy OnErrorStrategy, isPanic *bool, recoverValue interface{}, errorSink ...ErrorSink1) {
	if onErrorStrategy == recoverOnErrrorNone {
		if errp == nil {
			panic(NilError("errp"))
		}
	} else if errp == nil && errorSink == nil {
		panic(NilError("both errp and onError"))
	}
	var eSink ErrorSink1
	if len(errorSink) > 0 {
		eSink = errorSink[0]
	}

	// build aggregate error in err
	var err error
	if errp != nil {
		err = *errp
		// if onError is to be invoked multiple times,
		// and *errp contains an error,
		// invoke onError or Log to standard error
		if err != nil && onErrorStrategy == recoverOnErrrorMultiple {
			sendError(eSink, err) // invoke onError or parl.Log
		}
	}

	// consume recover()
	if recoverValue != nil {
		// update optional panic flag
		if isPanic != nil {
			*isPanic = true
		}
		annotation = getDeferredAnnotation(annotation, deferredAnnotation)
		var panicError = processRecoverValue(annotation, recoverValue, doRecoveryFrames)
		err = perrors.AppendError(err, panicError)
		if onErrorStrategy == recoverOnErrrorMultiple {
			sendError(eSink, panicError)
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
			sendError(eSink, err)
		}
	}
}

// getDeferredAnnotation returns an annotation string from a deferred annotation function literal
//   - annotation0: typically empty, can be string of some distinguishing property
//   - — “copy command i/o stderr” → “copy…: panic detected in…”
//   - deferredAnnotation: a thunk returning a pruntime.CodeLocation for where it is declared
//     on invocation providing package and function name enclosing the defer-statement invoking recover
//   - annotation:
//   - —“panic detected in os.Write:”
//   - — “recover from panic:”
//   - — “copy command i/o stderr: panic detected in os.Write:”
//   - — “copy command i/o stderr panic:”
func getDeferredAnnotation(annotation0 string, deferredAnnotation annotationLiteral) (annotation string) {
	if deferredAnnotation != nil {
		// execute the thunk returning the code location of the function literal declaration
		if da := deferredAnnotation(); da != nil {
			var cL = (*pruntime.CodeLocation)(da)
			// single word package name
			var packageName = cL.Package()
			// recoverDaPanic.func1: hosting function name and a derived name for the function literal
			var funcName = cL.FuncIdentifier()
			// function literals are anonymous functions with “.func1” suffix
			//	- remove the suffix
			if index := strings.LastIndex(funcName, "."); index != -1 {
				funcName = funcName[:index]
			}

			// annotation with code location
			if annotation0 != "" {
				annotation0 += "\x20: "
			}
			annotation = fmt.Sprintf("%spanic detected in %s.%s:",
				annotation0,
				packageName,
				funcName,
			)
			return
		}
	}

	// fixed annotation without code location
	if annotation0 != "" {
		annotation = annotation0 + " panic:"
		return
	}

	// annotation without code location
	// default annotation cannot be obtained
	//	- the deferred Recover function is invoked directly from rutine, eg. runtime.gopanic
	//	- therefore, use fixed string
	annotation = "recover from panic:"

	return
}

// sendError invokes an onError function or logs to standard error if onError is nil
func sendError(errorSink ErrorSink1, err error) {
	if errorSink != nil {
		errorSink.AddError(err)
		return
	}
	Log("Recover: %+v\n", err)
	Log("invokeOnError parl.recover %s", debug.Stack())
}

// processRecoverValue returns an error value with stack from annotation and panicValue
//   - annotation is non-empty annotation indicating code loction or action
//   - panicValue is non-nil value returned by built-in recover function
func processRecoverValue(annotation string, panicValue interface{}, frames int) (err error) {
	if frames < 0 {
		frames = 0
	}

	// if panicValue is an error with attached stack,
	// the panic detector will fail because
	// that innermost stack does not include panic recovery
	var hadPreRecoverStack bool
	if e, ok := panicValue.(error); ok {
		hadPreRecoverStack = errorglue.GetInnerMostStack(e) != nil
	}
	// ensure an error value is derived from panicValue
	err = perrors.Errorf("%s “%w”",
		annotation,
		ensureError(panicValue, frames+processRecoverFrames),
	)
	// make sure err has a post-recover() stack
	//	- this will allow the panic detector to succeed
	if hadPreRecoverStack {
		err = perrors.Stackn(err, frames)
	}

	return
}
