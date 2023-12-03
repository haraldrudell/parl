/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import "testing"

func TestSetElement(t *testing.T) {
	var value, name = 1, "nname"

	element := SetElement[int]{
		ValueV: value,
		Name:   name,
	}

	if element.Value() != value {
		t.Errorf("Value %d exp %d", element.Value(), value)
	}
	if element.String() != name {
		t.Errorf("String %q exp %q", element.String(), name)
	}

	if _, ok := any(&element).(Element[int]); !ok {
		t.Error("is not Element")
	}
}
