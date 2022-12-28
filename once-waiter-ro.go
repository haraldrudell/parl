/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"

	"github.com/haraldrudell/parl/perrors"
)

// OnceWaiterRO allows any number of threads to wait for a single next occurrence.
// OnceWaiterRO is a OnceWaiter without the Cancel method
type OnceWaiterRO struct {
	ow OnceWaiter
}

// NewOnceWaiter returns a channel that will send one item
// when the context cancels or immediately if the context was already canceled.
func NewOnceWaiterRO(onceWaiter *OnceWaiter) (onceWaiterRO *OnceWaiterRO) {
	if onceWaiter == nil {
		panic(perrors.NewPF("onceWaiter cannot be nil"))
	}
	return &OnceWaiterRO{ow: *onceWaiter}
}

func (ow *OnceWaiterRO) Ch() (ch <-chan struct{})       { return ow.ow.Ch() }
func (ow *OnceWaiterRO) Done() (done <-chan struct{})   { return ow.ow.Done() }
func (ow *OnceWaiterRO) Wait()                          { ow.ow.Wait() }
func (ow *OnceWaiterRO) DidOccur() (didOccur bool)      { return ow.ow.DidOccur() }
func (ow *OnceWaiterRO) Context() (ctx context.Context) { return ow.ow.Context() }
