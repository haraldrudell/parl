/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import "testing"

func Test_initG1ID(t *testing.T) {
	var id goEntityID
	_ = id
	id = *newGoEntityID(0)
	if id.id == 0 {
		t.Error("id 0")
	}
	if id.t.IsZero() {
		t.Error("t IsZero")
	}
	if id.creator.String() == "" {
		t.Error("creator empty")
	}
}
