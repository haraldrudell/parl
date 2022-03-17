/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

func TestAtomicBool(t *testing.T) {
	var b AtomicBool

	if b.IsTrue() {
		t.Errorf("default AtomicBool value is true")
	}
}
