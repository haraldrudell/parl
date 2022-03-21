/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"fmt"
	"testing"

	"github.com/haraldrudell/parl/errorglue"
)

func TestDumpChain(t *testing.T) {
	errorMessage := "an error"
	err := errors.New(errorMessage)
	err2 := Stack(err)
	expected := fmt.Sprintf("%T %T", err2, err)
	actual := errorglue.DumpChain(err2)
	if actual != expected {
		t.Errorf("DumpChain: %q expected: %q", actual, expected)
	}
}
