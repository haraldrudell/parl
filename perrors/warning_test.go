package perrors_test

import (
	"errors"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestIsWarning(t *testing.T) {
	err := errors.New("err")
	w := perrors.Warning(err) // mark as warning

	// outermost error is now the stack trace
	// *errorglue.errorStack *errorglue.WarningType *errors.errorString
	//t.Error(errorglue.DumpChain(w))

	actual := perrors.IsWarning(w)
	if !actual {
		t.Error("IsWarning broken")
	}
}
