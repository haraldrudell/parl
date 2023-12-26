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
	Add(goEntityID parl.GoEntityID, threadData *ThreadData)
	CascadeEnableTermination(delta int)
	ConsumeError(goError parl.GoError)
	GoDone(g parl.Go, err error)
	UpdateThread(goEntityID parl.GoEntityID, threadData *ThreadData)
	Context() (ctx context.Context)
}
