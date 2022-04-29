/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

// parl.Once is an observable sync.Once
// parl.Once is thread-safe and does not require initialization
// No thread will return from Once.Do until once.Do has completed
type Once struct {
	once   sync.Once
	isDone AtomicBool
}

func (o *Once) Do(f func()) {
	o.once.Do(func() {
		defer o.isDone.Set()
		f()
	})
}

func (o *Once) IsDone() (isDone bool) {
	return o.isDone.IsTrue()
}
