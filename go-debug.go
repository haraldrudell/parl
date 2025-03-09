/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/sets"

// controls SetDebug [NoDebug] [DebugPrint] [AggregateThread]
//   - [GoGroup.SetDebug] [SubGo.SetDebug] [SubGroup.SetDebug]
type GoDebug uint8

func (g GoDebug) String() (s string) {
	return goDebugSet.StringT(g)
}

// set providing striung values for GoDebug
var goDebugSet = sets.NewSet[GoDebug]([]sets.SetElement[GoDebug]{
	{ValueV: NoDebug, Name: "NoDebug"},
	{ValueV: DebugPrint, Name: "DebugPrint"},
	{ValueV: AggregateThread, Name: "AggregateThread"},
})
