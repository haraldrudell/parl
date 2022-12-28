/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"testing"
)

func Test_g1Factory(t *testing.T) {
	g1Group := GoGroupFactory.NewGoGroup(context.Background())
	if g1Group == nil {
		t.Error("G1Factory.NewG1 nil")
	}
}
