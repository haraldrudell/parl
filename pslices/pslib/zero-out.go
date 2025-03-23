/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslib

import "github.com/haraldrudell/parl/sets"

const (
	// [SliceAwayAppend] [SliceAwayAppend1] do zero-out obsolete slice elements
	//   - [SetLength] noZero
	DoZeroOut ZeroOut = iota
	// [SliceAwayAppend] [SliceAwayAppend1] do not zero-out obsolete slice elements
	//   - [SetLength] noZero
	NoZeroOut
)

// optional argument [NoZeroOut] to map and slice methods
type ZeroOut uint8

func (z ZeroOut) String() (s string) {
	return zeroOutSet.StringT(z)
}

var zeroOutSet = sets.NewSet[ZeroOut]([]sets.SetElement[ZeroOut]{
	{ValueV: DoZeroOut, Name: "DoZeroOut"},
	{ValueV: NoZeroOut, Name: "NoZeroOut"},
})
