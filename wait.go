/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type WaitDo struct {
	Waiter
}

func (w *WaitDo) IsExit() (isExit bool) {
	return w.IsZero()
}
