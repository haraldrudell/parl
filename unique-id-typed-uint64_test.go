/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

type T uint64

var generator UniqueIDTypedUint64[T]

func (t T) String() string { return generator.StringT(t) }
func TestUniqueIDTypedUint64(t *testing.T) {
	var v1 T = 1
	var v2 T = 2

	var v T

	if v = generator.ID(); v != v1 {
		t.Errorf("bad1 %d exp %d", v, v1)
	}
	if v = generator.ID(); v != v2 {
		t.Errorf("bad2 %d exp %d", v, v2)
	}
	t.Logf("type name via %%T: %T", v)

	// format %s has arg v of wrong type github.com/haraldrudell/parl.T
	t.Logf("value via %%s: %s", v)

	t.Logf("value via %%v: %v", v)
	var vNil T
	t.Log(generator.StringT(vNil))
	//t.Fail()
}
