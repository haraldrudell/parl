/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"testing"
)

func TestPanicDetector(t *testing.T) {
	//t.Fail()
	/*
		var detector panicDetector
		_ = detector

		// should execute without panic
		// retrieve all temporary data for post-mortem analysis
		detector = panicDetector{}
		panicCL, recoveryCL, stack := detector.getPanicData()

		// now check if it works
		stack1 := func() (s pruntime.StackSlice) {
			defer func() {
				recover()
				s = pruntime.NewStackSlice(0)
			}()
			panic(1)
		}()
		pd := panicDetectorOne
		panicDetectorOne = &detector
		defer func() {
			panicDetectorOne = pd
		}()
		isPanic, recoveryIndex, panicIndex := Indices(stack)
		if !isPanic {
			t.Error("isPanic false")
		}
		// recoveryIndex is 0, any value is ok. if isPanic is true, it did succeed
		if panicIndex == 0 {
			t.Error("panicIndex 0")
		}

		// dump out all data for forensics
		t.Logf("detector.getPanicData():\n"+
			"panicCL: the code location where detector invoked panic:\n%s\n"+
			"recoveryCL: the code location where detector recovered\n%s\n"+
			"panicLine: the stack trace line before a panic() frame:\n%s\n"+
			"deferLine: the stack frame after a recover() frame:\n%s\n"+
			"detector stack:%s\n\n",
			panicCL, recoveryCL,
			pd.runtimePanicFunctionLocation,
			pd.runtimeDeferInvokerLocation,
			stack,
		)
		t.Logf("Indices function: isPanic: %t recoveryIndex: %d panicIndex: %d stack:%s\n\n",
			isPanic, recoveryIndex, panicIndex, stack1)
	*/
}
