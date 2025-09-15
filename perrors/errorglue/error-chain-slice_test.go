/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/haraldrudell/parl/perrors/errorglue"
)

func TestErrorChainSlice(t *testing.T) {
	err := errors.New("x")
	err2 := fmt.Errorf("%w", err)
	errLen := 2

	var errs []error

	errs = errorglue.ErrorChainSlice(nil)
	if len(errs) != 0 {
		t.Errorf("errs len not 0: %d", len(errs))
	}

	errs = errorglue.ErrorChainSlice(err2)
	if len(errs) != errLen {
		t.Errorf("errs len not %d: %d", errLen, len(errs))
	}
}
