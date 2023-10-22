/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import (
	"github.com/haraldrudell/parl/pruntime"
)

// OutputInvoker provides Invoke function for specific code location
type OutputInvoker struct {
	// the code location appended to debug print
	codeLocation *pruntime.CodeLocation
	// [LogInstance.invokeOutput]
	invokeOutput func(s string)
}

// NewOutputInvoker returns a [LogInstance.invokeOutput] function for a specific code location
func NewOutputInvoker(
	codeLocation *pruntime.CodeLocation,
	invokeOutput func(s string),
) (outputInvoker *OutputInvoker) {
	return &OutputInvoker{
		codeLocation: codeLocation,
		invokeOutput: invokeOutput,
	}
}

// Invoke invokes the [LogInstance.invokeOutput] function for a stored code location
func (o *OutputInvoker) Invoke(format string, a ...any) {
	o.invokeOutput(
		pruntime.AppendLocation(
			Sprintf(format, a...), // collapse format and a to single string
			o.codeLocation,
		))
}
