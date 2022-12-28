/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"testing"
)

func TestGoWaitGroup(t *testing.T) {
	var g0WaitGroup *goWaitGroup
	_ = g0WaitGroup

	g0WaitGroup = newGoWaitGroup(context.Background())
	g0WaitGroup.Wait()
}
