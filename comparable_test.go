/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

type valueReceiverX int

func (a valueReceiverX) Cmp(b valueReceiverX) (result int) {
	if a > b {
		return 1
	} else if a < b {
		return -1
	}
	return 0
}

func TestComparable(t *testing.T) {
	v1 := 1
	v2 := 2
	var a Comparable[valueReceiverX] = valueReceiverX(v1)
	var b = valueReceiverX(v2)
	exp1 := -1

	var actual int
	_ = actual

	actual = a.Cmp(b)
	if actual != exp1 {
		t.Errorf("cmp1 bad %d exp %d", actual, exp1)
	}
}

type pointerReceiverX struct{ v int }

func (a *pointerReceiverX) Cmp(b *pointerReceiverX) (result int) {
	if a.v > b.v {
		return 1
	} else if a.v < b.v {
		return -1
	}
	return 0
}

func TestComparableP(t *testing.T) {
	var a Comparable[*pointerReceiverX] = &pointerReceiverX{1}
	var b = &pointerReceiverX{2}
	exp1 := -1

	var actual int
	_ = actual

	actual = a.Cmp(b)
	if actual != exp1 {
		t.Errorf("cmp1 bad %d exp %d", actual, exp1)
	}
}
