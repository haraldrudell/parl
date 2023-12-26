/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

// goParent are the methods provided to a Go thread by its parent GoGroup thread-group
type goParent interface {
	ConsumeError(goError parl.GoError)
	FromGoGo() (g parl.Go)
	FromGoSubGo(onFirstFatal ...parl.GoFatalCallback) (g parl.SubGo)
	FromGoSubGroup(onFirstFatal ...parl.GoFatalCallback) (g parl.SubGroup)
	GoDone(g parl.Go, err error)
	UpdateThread(goEntityID parl.GoEntityID, threadData *ThreadData)
	Cancel()
	Context() (ctx context.Context)
}
