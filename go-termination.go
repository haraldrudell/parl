/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/sets"

// controls SetDebug [NoDebug] [DebugPrint] [AggregateThread]
//   - [GoGroup.SetDebug] [SubGo.SetDebug] [SubGroup.SetDebug]
type GoTermination uint8

func (g GoTermination) String() (s string) {
	return goTerminationSet.StringT(g)
}

// set providing string values for GoTermination
var goTerminationSet = sets.NewSet[GoTermination]([]sets.SetElement[GoTermination]{
	{ValueV: AllowTermination, Name: "AllowTermination"},
	{ValueV: PreventTermination, Name: "PreventTermination"},
})
