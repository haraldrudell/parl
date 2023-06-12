/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"testing"
)

func TestContextID(t *testing.T) {
	var c1 context.Context = context.Background()
	var c2 context.Context = c1
	var s1, s2 = ContextID(c1), ContextID(c2)
	if s1 != s2 {
		t.Errorf("C1_%s exp\nC2_%s", s1, s2)
	}
}
