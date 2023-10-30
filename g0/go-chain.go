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

func GoChain(g parl.GoGen) (s string) {
	for {
		var s0 = GoNo(g)
		if s == "" {
			s = s0
		} else {
			s += "—" + s0
		}
		if g == nil {
			return
		} else if g = Parent(g); g == nil {
			return
		}
	}
}

func Parent(g parl.GoGen) (parent parl.GoGen) {
	switch g := g.(type) {
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

func GoNo(g parl.GoGen) (goNo string) {
	switch g1 := g.(type) {
	case *Go:
		goNo = "Go" + g1.id.String() + ":" + g1.GoID().String()
	case *GoGroup:
		if !g1.hasErrorChannel {
			goNo = "SubGo"
		} else if g1.parent != nil {
			goNo = "SubGroup"
		} else {
			goNo = "GoGroup"
		}
		goNo += g1.id.String()
	case nil:
		goNo = "nil"
	default:
		goNo = fmt.Sprintf("?type:%T", g)
	}
	return
}
