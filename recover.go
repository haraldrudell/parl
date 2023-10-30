/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	panicString              = ": panic:"
	recover2OnErrrorOnce     = false
	recover2OnErrrorMultiple = true
)

const (
	// counts the stack-frame in [parl.processRecover]
	processRecoverFrames = 1
	// counts the stack-frame of [parl.doRecovery] and [parl.Recover] or [parl.Recover2]
	//	- but for panic detectpr to work, there must be one frame after
	//		runtime.gopanic, so remove one frame
	doRecoveryFrames = 2 - 1
)

// OnError is a function that receives error values from an errp error pointer or a panic
type OnError func(err error)

// NoOnError is used with Recover and Recover2 to silence the default error logging
func NoOnError(err error) {}

var noIsPanic *bool

// Recover recovers from panic invoking onError exactly once with an aggregate error value
//   - annotation may be empty, errp and onError may be nil
//   - errors in *errp and panic are aggregated into a single error value
//   - if onError non-nil, the function is invoked once with the aggregate error
//   - if onError nil, the aggregate error is logged to standard error
//   - if onError is [Parl.NoOnErrror], logging is suppressed
//   - if errp is non-nil, it is updated with the aggregate error
//   - if annotation is empty, a default annotation is used for the immediate caller of Recover
func Recover(annotation string, errp *error, onError OnError) {
	doRecovery(annotation, nil, errp, onError, recover2OnErrrorOnce, noIsPanic, recover())
}

// Recover2 recovers from panic invoking onError for any eror in *errp and any panic
//   - annotation may be empty, errp and onError may be nil
//   - if onError non-nil, the function is invoked with any error in *errp and any panic
//   - if onError nil, the errors are logged to standard error
//   - if onError is [Parl.NoOnErrror], logging is suppressed
//   - if errp is non-nil, it is updated with an aggregate error
//   - if annotation is empty, a default annotation is used for the immediate caller of Recover
func Recover2(annotation string, errp *error, onError OnError) {
	doRecovery(annotation, nil, errp, onError, recover2OnErrrorMultiple, noIsPanic, recover())
}

// doRecovery implements recovery ffor Recovery andd Recovery2
func doRecovery(annotation string, deferredAnnotation func() DA, errp *error, onError OnError, multiple bool, isPanic *bool, recoverValue interface{}) {

	// build aggregate error in err
	var err error

	// if onError is to be invoked multiple times,
	// and *errp contains an error,
	// invoke onError or Log to standard error
	if errp != nil {
		if err = *errp; err != nil && multiple {
			invokeOnError(onError, err) // invokee onError or parl.Log
		}
	}

	// consume recover()
	if recoverValue != nil {
		if isPanic != nil {
			*isPanic = true
		}
		if annotation == "" {
			if deferredAnnotation != nil {
				if da := deferredAnnotation(); da != nil {
					var cL = (*pruntime.CodeLocation)(da)
					// single wword package name
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
		}
		e := processRecover(annotation, recoverValue, doRecoveryFrames)
		if multiple {
			invokeOnError(onError, e)
		} else {
			err = perrors.AppendError(err, e)
		}
	}

	// if err now contains any error
	//	- write bacxk to non-nil errp
	//	- if not multiple, invoke onErorr or Log the aggregate error
	if err != nil {
		if errp != nil && *errp != err {
			*errp = err
		}
		if !multiple {
			invokeOnError(onError, err)
		}
	}
}

// invokeOnError invokes an onError function or logs to standard error if onError is nil
func invokeOnError(onError OnError, err error) {
	if onError != nil {
		onError(err)
		return
	}
	Log("Recover: %+v\n", err)
}

// processRecover ensures non-nil result to be error with Stack
//   - annotation is non-empty annotation indicating code loction or action
//   - panicValue is non-nil value returned by built-in recover function
func processRecover(annotation string, panicValue interface{}, frames int) (err error) {
	if frames < 0 {
		frames = 0
	}
	return perrors.Errorf("%s “%w”",
		annotation,
		ensureError(panicValue, frames+processRecoverFrames),
	)
}
