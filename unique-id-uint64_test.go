/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

func TestUniqueIDUint64(t *testing.T) {
	var v1 uint64 = 1
	var v2 uint64 = 2

	var v uint64

	var generator UniqueIDUint64
	if v = generator.ID(); v != v1 {
		t.Errorf("bad1 %d exp %d", v, v1)
	}
	if v = generator.ID(); v != v2 {
		t.Errorf("bad2 %d exp %d", v, v2)
	}

	// v.String undefined (type uint64 has no field or method String)
	//v.String()

	// conversion from uint64 to string yields a string of one rune, not a string of digits
	// (did you mean fmt.Sprint(x)?)
	//_ = string(v)

	// format %s has arg v of wrong type uint64, see also https://pkg.go.dev/fmt#hdr-Printing
	//t.Logf("print using %%s: %s", v)

	// print using %d: 2
	t.Logf("print using %%d: %d", v)

	// The default format for %v is:
	// uint, uint8 etc.: %d, %#x if printed with %#v
	// print using %v: 2
	t.Logf("print using %%v: %v", v)

	//t.Fail()
}
