/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

func TestUniqueID(t *testing.T) {
	//t.Fail()
	expect1 := "1"
	expect2 := "2"

	type myType string
	var generator UniqueID[myType]
	var actual myType

	actual = generator.ID()

	// parl.myType "1"
	t.Logf("%T %[1]q", actual)
	if actual != myType(expect1) {
		t.Errorf("actual 1: %q exp %q", actual, expect1)
	}

	actual = generator.ID()

	if actual != myType(expect2) {
		t.Errorf("actual 1: %q exp %q", actual, expect2)
	}
}
