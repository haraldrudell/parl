/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/sets"

const (
	// shared lazy-created thread that never exits
	SlowDefault SlowType = iota
	// dedicated lazy-created thread that never exits
	SlowOwnThread
	// dedicated lazy-created thread that exits whenever monitored invocations is zero
	SlowShutdownThread
)

// [SlowDefault] [SlowOwnThread] [SlowShutdownThread]
type SlowType uint8

func (st SlowType) String() (s string) { return slowTypeSet.StringT(st) }

// [SlowType] set
var slowTypeSet = sets.NewSet[SlowType]([]sets.SetElement[SlowType]{
	{ValueV: SlowDefault, Name: "sharedThread"},
	{ValueV: SlowOwnThread, Name: "ownThread"},
	{ValueV: SlowShutdownThread, Name: "shutdownThread"},
})
