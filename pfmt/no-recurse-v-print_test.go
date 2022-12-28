/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package pfmt

import (
	"fmt"
	"testing"
)

type typeT struct{}

var _ fmt.Stringer = &typeT{}

func (valueT typeT) String() (s string) {
	return f(valueT)
}

func f(valueT typeT) (s string) {
	return fmt.Sprint(valueT)
}

func TestNoRecurseVPrint(t *testing.T) {

	var valueT typeT

	// runtime: goroutine stack exceeds 1000000000-byte limit
	// fatal error: stack overflow
	//_ = fmt.Sprint(valueT)

	NoRecurseVPrint(valueT)
}
