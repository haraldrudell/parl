/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

type WaitAction struct {
	ID     ThreadID
	Loc    pruntime.CodeLocation
	IsDone bool
	Delta  int
}

func NewWaitAction(skipFrames int, delta int, isDone bool) (waitAction *WaitAction) {
	if newStack == nil {
		panic(perrors.NewPF("pdebug has not invoked ImportNewStack"))
	}
	stack := newStack(skipFrames)
	return &WaitAction{
		ID:     stack.ID(),
		Loc:    *stack.Frames()[0].Loc(),
		IsDone: isDone,
		Delta:  delta,
	}
}

func (wa *WaitAction) String() (s string) {
	if wa.IsDone {
		s = "done"
	} else {
		s = Sprintf("%+d", wa.Delta)
	}
	return s + "\x20" + wa.ID.String() + "\x20" + wa.Loc.Short()
}
