/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"testing"
)

const (
	one Elem = iota + 1
	two
)

type Elem uint8

func TestNewElements(t *testing.T) {
	oneName := "one"
	twoName := "two"

	var set Set[Elem]
	var oneString string

	set = NewSet(NewElements[Elem](
		[]SetElement[Elem]{
			{one, oneName},
			{two, twoName},
		}))
	if oneString = set.StringT(one); oneString != oneName {
		t.Errorf("StringT(one): %q exp %q", oneString, oneName)
	}
}
