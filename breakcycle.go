/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/breakcycle"
	"github.com/haraldrudell/parl/perrors"
)

// if no go statements are executed prior to goid package being
// initialized, then this lock and the function call are not required.
// a go statement is happens before.
//var lock sync.Mutex
var newStack func(skipFrames int) (stack Stack)

var _ = func() (i int) {
	breakcycle.ParlGoidImport(setNewStack)
	return
}()

/*
func getNewStack() (newStackFn func(skipFrames int) (stack Stack)) {
	lock.Lock()
	defer lock.Unlock()

	return newStack
}
*/
func setNewStack(v interface{}) {
	var ok bool
	if newStack, ok = v.(func(skipFrames int) (stack Stack)); !ok {
		panic(perrors.Errorf("setNewStack: v bad type: %T %[1]v", v))
	}
}
