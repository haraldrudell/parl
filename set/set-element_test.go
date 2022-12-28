/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package set

import "testing"

func TestSetElement(t *testing.T) {
	value := 1
	name := "nname"

	element := Element[int]{value, name}

	if element.Value() != value {
		t.Errorf("Value %d exp %d", element.Value(), value)
	}
	if element.String() != name {
		t.Errorf("String %q exp %q", element.String(), name)
	}
}
