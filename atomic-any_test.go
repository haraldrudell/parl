package parl_test

import (
	"io"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestAtomicAny_Load(t *testing.T) {
	var atomicCloser parl.AtomicAny[io.Closer]

	var closer io.Closer = atomicCloser.Load()
	if closer != nil {
		t.Error("closer not nil")
	}
}
