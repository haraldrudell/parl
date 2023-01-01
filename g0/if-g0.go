/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

type goGroupParent interface {
	Add(id GoEntityID, threadData *ThreadData)
	ConsumeError(goError parl.GoError)
	GoDone(g0 parl.Go, err error)
	UpdateThread(goEntityID GoEntityID, threadData *ThreadData)
	Context() (ctx context.Context)
}

// goParent are the methods provided to a Go thread by its parent GoGroup thread-group
type goParent interface {
	ConsumeError(goError parl.GoError)
	Go() (g1 parl.Go)
	SubGo(onFirstFatal ...parl.GoFatalCallback) (g0 parl.SubGo)
	SubGroup(onFirstFatal ...parl.GoFatalCallback) (g0 parl.SubGroup)
	GoDone(g0 parl.Go, err error)
	UpdateThread(goEntityID GoEntityID, threadData *ThreadData)
	Cancel()
}

type goImpl interface {
	G0ID() (id GoEntityID)
	ThreadData() (threadData *ThreadData)
}

type goParentArg interface {
	goParent
	Context() (ctx context.Context)
}