/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psync

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
	"github.com/haraldrudell/parl/pruntime"
)

type WaitAction struct {
	ID     parl.ThreadID
	Loc    pruntime.CodeLocation
	IsDone bool
	Delta  int
}

func NewWaitAction(skipFrames int, delta int, isDone bool) (waitAction *WaitAction) {
	stack := goid.NewStack(skipFrames)
	return &WaitAction{
		ID:     stack.ID,
		Loc:    stack.Frames[0].CodeLocation,
		IsDone: isDone,
		Delta:  delta,
	}
}

func (wa *WaitAction) String() (s string) {
	if wa.IsDone {
		s = "done"
	} else {
		s = parl.Sprintf("%+d", wa.Delta)
	}
	return s + "\x20" + wa.ID.String() + "\x20" + wa.Loc.Short()
}
