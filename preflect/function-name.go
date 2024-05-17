/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"reflect"
	"runtime"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

// FuncName returns information on a function value
//   - funcOrMethod: non-nil function value such as:
//   - — a top-level function: “func main() {”
//   - — a method: “errors.New("").Error”
//   - — an anonymous function: “var f = func() {}”
//     returned like “SomeFunc.func1”
//   - cL is a function-describing structure of basic types only
//   - — [pruntime.CodeLocation.FuncIdentifier] returns the function name identifier,
//     a single word
//   - error is returned if:
//   - — funcOrMethod is nil or not a function value
//   - — function value retrieval failed in runtime
func FuncName(funcOrMethod any) (cL *pruntime.CodeLocation, err error) {

	// funcOrMethod cannot be nil
	if funcOrMethod == nil {
		err = parl.NilError("funcOrMethod")
		return
	}

	// funcOrMethod must be underlying type func
	var reflectValue = reflect.ValueOf(funcOrMethod)
	if reflectValue.Kind() != reflect.Func {
		err = perrors.ErrorfPF("funcOrMethod not func: %T", funcOrMethod)
		return
	}

	// get func name, “func1” for anonymous
	var runtimeFunc = runtime.FuncForPC(reflectValue.Pointer())
	if runtimeFunc == nil {
		err = perrors.NewPF("runtime.FuncForPC returned nil")
		return
	}
	cL = pruntime.CodeLocationFromFunc(runtimeFunc)

	return
}
