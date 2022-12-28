/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"errors"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestG1(t *testing.T) {
	err := errors.New("x")

	var g1Group parl.GoGroup
	var g1 parl.Go
	var g1Impl *Go
	var subGroup parl.SubGroup
	var subGo parl.SubGo

	g1Group = NewGoGroup(context.Background())

	// newG1
	g1 = g1Group.Go()
	g1Impl = g1.(*Go)
	if g1Impl.isTerminated.IsTrue() {
		t.Error("G1 terminated")
	}

	// Register SubGo SubGroup AddError Done
	g1.Register()
	subGroup = g1.SubGroup()
	_ = subGroup
	subGo = g1.SubGo()
	_ = subGo
	g1.AddError(err)
	g1.Done(&err)

	// String
	t.Log(g1.String())
	//t.Fail()
}
