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
	Add(id parl.GoEntityID, threadData *ThreadData)
	CascadeEnableTermination(delta int)
	ConsumeError(goError parl.GoError)
	GoDone(g0 parl.Go, err error)
	UpdateThread(goEntityID parl.GoEntityID, threadData *ThreadData)
	Context() (ctx context.Context)
}

// goParent are the methods provided to a Go thread by its parent GoGroup thread-group
type goParent interface {
	ConsumeError(goError parl.GoError)
	FromGoGo() (g1 parl.Go)
	FromGoSubGo(onFirstFatal ...parl.GoFatalCallback) (g0 parl.SubGo)
	FromGoSubGroup(onFirstFatal ...parl.GoFatalCallback) (g0 parl.SubGroup)
	GoDone(g0 parl.Go, err error)
	UpdateThread(goEntityID parl.GoEntityID, threadData *ThreadData)
	Cancel()
	Context() (ctx context.Context)
}
