/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"sync"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestThreadLogger(t *testing.T) {
	var g0Group parl.GoGroup
	var wg sync.WaitGroup

	g0Group = NewGoGroup(context.Background())
	wg.Add(1)
	ThreadLogger(g0Group, &wg)
	wg.Wait()
}
