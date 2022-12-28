/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"

	"github.com/haraldrudell/parl/perrors"
)

// OnceWaiter allows any number of threads to wait for a single next occurrence.
//   - the occurrence is trigger in any of two ways:
//   - — a parent context may be passed in that on cancel triggers the wait
//   - — the Cancel method is invoked
//   - Ch returns a channel that sends one item on the occurrence
//     but never closes
//   - Done returns a channel that closes on the occurrence happens, similar
//     to a context
//   - Wait awaits the occurrence
//   - a did-occurer object can be obtained that returns true once the cycle
//     trigs.
//   - a context can be obtained that cancels on the next trig
//   - the cycles can be permanently canceled or trigged and rearmed
type OnceWaiter struct {
	ctx context.Context
}

// NewOnceWaiter returns a channel that will send one item
// when the context cancels or immediately if the context was already canceled.
func NewOnceWaiter(ctx context.Context) (onceReceiver *OnceWaiter) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	}
	return &OnceWaiter{
		ctx: NewCancelContext(ctx),
	}
}

// Ch returns a channel that will send one item on the occurrence.
//   - the channel will not send anything else
//   - the channel never closes.
func (ow *OnceWaiter) Ch() (ch <-chan struct{}) {
	c := make(chan struct{}, 1)
	go onceWaiterSender(ow.ctx.Done(), c)
	return c
}

// Done returns a channel that will close on the occurrence.
//   - Done is similar to the Done method of a context
func (ow *OnceWaiter) Done() (done <-chan struct{}) {
	return ow.ctx.Done()
}

// Wait waits until the ocurrence
func (ow *OnceWaiter) Wait() {
	done := ow.Done()
	<-done
}

// DidOccur returns true when the occurrence has taken place
func (ow *OnceWaiter) DidOccur() (didOccur bool) {
	return ow.ctx.Err() != nil
}

// Context returns a context that cancels on the occurrence
func (cw *OnceWaiter) Context() (ctx context.Context) {
	return NewCancelContext(cw.ctx)
}

// Cancel triggers the occurrence
func (ow *OnceWaiter) Cancel() {
	invokeCancel(ow.ctx)
}

func onceWaiterSender(done <-chan struct{}, ch chan<- struct{}) {
	Recover(Annotation(), nil, Infallible) // panic prints to stderr

	<-done
	ch <- struct{}{}
}
