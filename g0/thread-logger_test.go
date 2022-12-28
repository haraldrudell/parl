/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestThreadLogger(t *testing.T) {
	var g0Group parl.GoGroup
	_ = g0Group

	g0Group = NewGoGroup(context.Background())
	ThreadLogger(g0Group, nil)
}
