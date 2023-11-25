/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/sets"

const (
	ITParallelism CBReason = iota + 1
	ITLatency
)

// CBReason explains to consumer why the callback was invoked
//   - ITParallelism ITLatency
type CBReason uint8

func (r CBReason) String() (s string) {
	return cbReasonSet.StringT(r)
}

var cbReasonSet = sets.NewSet(sets.NewElements[CBReason](
	[]sets.SetElement[CBReason]{
		{ValueV: ITParallelism, Name: "max parallel"},
		{ValueV: ITLatency, Name: "slowest"},
	}))
