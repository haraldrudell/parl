/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/sets"

const (
	// [parl.InvocationTimer] callback due to increased parallelism
	ITParallelism CBReason = iota + 1
	// [parl.InvocationTimer] callback due to increased latency
	ITLatency
)

// CBReason explains to consumer why [parl.InvocationTimer] invoked the callback
//   - ITParallelism ITLatency
type CBReason uint8

func (r CBReason) String() (s string) {
	return cbReasonSet.StringT(r)
}

// cbReasonSet translates CBReason to string
var cbReasonSet = sets.NewSet(sets.NewElements[CBReason](
	[]sets.SetElement[CBReason]{
		{ValueV: ITParallelism, Name: "max parallel"},
		{ValueV: ITLatency, Name: "slowest"},
	}))
