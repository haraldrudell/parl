/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

func TestAtomicBool(t *testing.T) {
	var b AtomicBool

	if b.IsTrue() {
		t.Error("default AtomicBool value is true")
	}

	if !b.Set() {
		t.Error("Set returned false")
	}
	if b.Set() {
		t.Error("Set returned true second time")
	}
	if !b.Clear() {
		t.Error("Clear returned true second time")
	}
}
