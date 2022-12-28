/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"testing"

	"github.com/haraldrudell/parl"
)

func TestOrdered(t *testing.T) {
	v1 := 2
	v2 := 4

	var ordered parl.Ordered[int]
	var ordered2 parl.Ordered[int]
	var index int

	ordered = NewOrdered[int]()

	// check duplicates
	ordered.Insert(v1)
	ordered.Insert(v1)
	if ordered.Length() != 1 {
		t.Errorf("Length not 1: %d", ordered.Length())
	}

	index = ordered.Index(v1)
	if index != 0 {
		t.Errorf("Index not 0: %d", index)
	}
	index = ordered.Index(v2)
	if index != -1 {
		t.Errorf("Index not -1: %d", index)
	}

	ordered2 = ordered.Clone()
	if ordered2.Length() != 1 {
		t.Errorf("Length2 not 1: %d", ordered.Length())
	}

	ordered.Delete(v2)
	ordered.Delete(v1)

}
