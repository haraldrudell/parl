/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/pslices/pslib"
)

const (
	// [SliceAwayAppend] [SliceAwayAppend1] do zero-out obsolete slice elements
	//   - [SetLength] noZero
	DoZeroOut = pslib.DoZeroOut
	// [SliceAwayAppend] [SliceAwayAppend1] do not zero-out obsolete slice elements
	//   - [SetLength] noZero
	NoZeroOut = pslib.NoZeroOut
)

// optional argument [NoZeroOut] [DoZeroOut] to map and slice methods
type ZeroOut = pslib.ZeroOut
