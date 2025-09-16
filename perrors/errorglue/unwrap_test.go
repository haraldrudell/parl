/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/haraldrudell/parl/perrors/errorglue"
)

func TestUnwrap(t *testing.T) {
	var (
		err0       = errors.New("zero")
		err1       = errors.New("one")
		errRelated = appendError(err0, err1)
		errWrapped = fmt.Errorf("%w", err0)
		errJoined  = errors.Join(err0, err1)
	)
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		err       error
		resultErr error
	}{
		{"nil", nil, nil},
		{"empty chain", err0, nil},
		{"perrors.AppendError", errRelated, err0},
		{"fmt.Errorf", errWrapped, err0},
		{"errors.Join", errJoined, err0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr, _, _ := errorglue.Unwrap(tt.err)
			if gotErr != tt.resultErr {
				t.Errorf("Unwrap() failed: %v exp %v", gotErr, tt.resultErr)
			}
		})
	}
}

func appendError(err error, err2 error) (e error) {
	if err2 == nil {
		e = err // err2 is nil, return is err, possibly nil
	} else if err == nil {
		e = err2 // err is nil, return is non-nil err2
	} else {
		e = errorglue.NewRelatedError(err, err2) // both non-nil
	}
	return
}
