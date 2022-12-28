/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

const g1Gropupg0StackFranmes = 1

// goWaitGroup is a promotable private field implementing
// Wait().
//   - WaitGroup is the core of a thread-group determining when it will terminate
type goWaitGroup struct {
	wg        parl.WaitGroup // Wait()
	goContext                // Cancel() Context()
}

// newGoWaitGroup initializes a Go wait-group
func newGoWaitGroup(ctx context.Context) (g0WaitGroup *goWaitGroup) {
	return &goWaitGroup{goContext: *newGoContext(ctx)}
}

func (g0 *goWaitGroup) Wait() { g0.wg.Wait() }
