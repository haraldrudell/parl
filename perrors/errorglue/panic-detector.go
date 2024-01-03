/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

const (
	runtimeGopanic = "runtime.gopanic"
)

type panicDetector struct {
	runtimeDeferInvokerLocation  string
	runtimePanicFunctionLocation string
}

// panicDetectorOne is a static value facilitating panic detection for this runtime.
// panicDetectorOne is created during package initialization and is
// therefore thread-safe.
//   - panicDetectorOne is used by [errorglue.Indices]
var panicDetectorOne = func() (pd *panicDetector) {
	p := panicDetector{
		runtimeDeferInvokerLocation:  runtimeGopanic,
		runtimePanicFunctionLocation: runtimeGopanic,
	}
	//p.getPanicData()
	return &p
}()

func PanicDetectorValues() (deferS, panicS string) {
	deferS = panicDetectorOne.runtimeDeferInvokerLocation
	panicS = panicDetectorOne.runtimePanicFunctionLocation
	return
}

// test panicDetector on this runtime
/*
221221 disabled
var _ = func() (i int) {
	defer func() {
		recover()
		isPanic, recoveryIndex, panicIndex := Indices(pruntime.NewStackSlice(0))
		if !isPanic || panicIndex == 0 {
			panic(fmt.Errorf("panicDetector failure: isPanic: %t recoveryIndex: %d panicIndex: %d",
				isPanic, recoveryIndex, panicIndex))
		}
	}()
	panic(1)
}()

func (pd *panicDetector) getPanicData() (
	p, r *pruntime.CodeLocation, s pruntime.StackSlice) {

	// get code locations and stack for a panic
	return pd.findRuntimeFrames(pd.recoveryFunction())
}

func (pd *panicDetector) findRuntimeFrames(panicCL, recoveryCL *pruntime.CodeLocation, stack pruntime.StackSlice) (
	p, r *pruntime.CodeLocation, s pruntime.StackSlice) {
	p = panicCL
	r = recoveryCL
	s = stack
	found := 0
	for i, stackFrame := range stack {

		// check if this is pd.panicFunction
		if stackFrame.FuncName == panicCL.FuncName {
			// runtime’s panic function is the frame before it
			if i > 0 {
				pd.runtimePanicFunctionLocation = stack[i-1].FuncLine()
				found++
				if found == 2 {
					break
				}
			}
		}

		// check if this is pd.recoveryFunction
		if stackFrame.FuncName == recoveryCL.FuncName {
			// runtime’s defer invoker is the next frame
			if i+1 < len(stack) {
				pd.runtimeDeferInvokerLocation = stack[i+1].FuncLine()
				found++
				if found == 2 {
					break
				}
			}
		}
	}

	if found != 2 {
		panic(errors.New("failed to find runtime frames"))
	}
	return
}

func (pd *panicDetector) recoveryFunction() (panicCL, recoveryCL *pruntime.CodeLocation, stack pruntime.StackSlice) {
	type t int

	defer func() {
		recoveryValue := recover()
		if _, ok := recoveryValue.(t); !ok {
			panic(fmt.Errorf("bad recovery value: %T %#[1]v", recoveryValue))
		}
		recoveryCL = pruntime.NewCodeLocation(0)
		stack = pruntime.NewStackSlice(0)
	}()

	var v t
	pd.panicFunction(v, &panicCL)
	return
}

func (pd *panicDetector) panicFunction(a any, panicCLp **pruntime.CodeLocation) {
	*panicCLp = pruntime.NewCodeLocation(0)
	panic(a)
}
*/
