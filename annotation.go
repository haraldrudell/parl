/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	// counts the stack-frame of [parl.Annotation]
	parlAnnotationFrames = 1
	// counts the stack-frame of [parl.getAnnotation]
	getAnnotationFrames = 1
)

// Annotation provides a default recovered-panic code annotation
//   - “Recover from panic in mypackage.MyFunc”
//   - [base package].[function]: "mypackage.MyFunc"
//   - Annotation is expensive at 0.40 parallel mutex Lock/Unlock:
//   - — consider using deferring methods [parl.RecoverErr] [parl.RecoverDA] [parl.RecoverDA2]
//   - — consider annotation-free [parl.PanicToErr] at less functionality
//   - if annotation is used in defer for every function invocation, that costs 470 ns
func Annotation() (a string) {
	return getAnnotation(parlAnnotationFrames)
}

// getAnnotation provides a default reovered-panic code getAnnotation
//   - frames = 0 means immediate caller of getAnnotation
//   - “Recover from panic in mypackage.MyFunc”
//   - [base package].[function]: "mypackage.MyFunc"
func getAnnotation(frames int) (a string) {
	if frames < 0 {
		frames = 0
	}
	return fmt.Sprintf("Recover from panic in %s:",
		pruntime.NewCodeLocation(frames+getAnnotationFrames).PackFunc(),
	)
}
