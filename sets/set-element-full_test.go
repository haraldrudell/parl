/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import "testing"

func TestSetElementFull(t *testing.T) {
	var value, name, full = 1, "nname", "ffull"

	var element = SetElementFull[int]{
		ValueV: value,
		Name:   name,
		Full:   full,
	}

	if element.Value() != value {
		t.Errorf("Value %d exp %d", element.Value(), value)
	}
	if element.Description() != full {
		t.Errorf("Description %q exp %q", element.Description(), full)
	}
	if element.String() != name {
		t.Errorf("String %q exp %q", element.String(), name)
	}

	if _, ok := any(&element).(ElementDescription); !ok {
		t.Error("is not ElementDescription")
	}
}
