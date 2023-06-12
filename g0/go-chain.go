/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"fmt"

	"github.com/haraldrudell/parl"
)

func GoChain(g0 parl.GoGen) (s string) {
	for {
		var s0 = GoNo(g0)
		if s == "" {
			s = s0
		} else {
			s += "—" + s0
		}
		if g0 == nil {
			return
		} else if g0 = Parent(g0); g0 == nil {
			return
		}
	}
}

func Parent(g0 parl.GoGen) (parent parl.GoGen) {
	switch g := g0.(type) {
	case *Go:
		parent = g.goParent.(parl.GoGen)
	case *GoGroup:
		if p := g.parent; p != nil {
			parent = p.(parl.GoGen)
		}
	}
	return
}

func ContextID(ctx context.Context) (contextID string) {
	return fmt.Sprintf("%x", parl.Uintptr(ctx))
}

func GoNo(g0 parl.GoGen) (goNo string) {
	switch g := g0.(type) {
	case *Go:
		goNo = "Go" + g.id.String() + ":" + g.GoID().String()
	case *GoGroup:
		if g.hasErrorChannel.IsFalse() {
			goNo = "SubGo"
		} else if g.parent != nil {
			goNo = "SubGroup"
		} else {
			goNo = "GoGroup"
		}
		goNo += g.id.String()
	case nil:
		goNo = "nil"
	default:
		goNo = fmt.Sprintf("?type:%T", g0)
	}
	return
}
