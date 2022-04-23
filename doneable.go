/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type DoneableNil struct{}

func EnsureDoneable(done Doneable) (done2 Doneable) {
	if done != nil {
		done2 = done
		return
	}
	return &DoneableNil{}
}

func (de *DoneableNil) Add(delta int) {}
func (de *DoneableNil) Done()         {}
