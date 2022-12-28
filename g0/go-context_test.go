/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"testing"
)

func TestGoContext(t *testing.T) {
	type X struct {
		goContext
	}
	x := &X{
		goContext: *newGoContext(context.Background()),
	}
	x.Context()
	x.Cancel()
}
