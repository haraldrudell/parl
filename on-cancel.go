/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "context"

// OnCancel invokes fn when work done on behalf of context ctx
// should be canceled
func OnCancel(fn func(), ctx context.Context) {
	go onCancelThread(fn, ctx.Done())
}

func onCancelThread(fn func(), done <-chan struct{}) {
	Recover(Annotation(), nil, Infallible)
	<-done
	fn()
}
