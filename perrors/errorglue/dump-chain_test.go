/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"testing"
)

func TestDumpChain(t *testing.T) {
	var (
		singleErr    = errors.New("x")
		expSingleErr = fmt.Sprintf("%T", singleErr)
		fmtErrorf    = fmt.Errorf("%w", singleErr)
		expFmtErrorf = fmt.Sprintf("%T %s", fmtErrorf, expSingleErr)
	)

	type args struct {
		err error
	}
	tests := []struct {
		name          string
		args          args
		wantTypeNames string
	}{
		{"nil", args{nil}, ""},
		{"single error", args{singleErr}, expSingleErr},
		{"fmt.Errorf", args{fmtErrorf}, expFmtErrorf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTypeNames := DumpChain(tt.args.err); gotTypeNames != tt.wantTypeNames {
				t.Errorf("DumpChain() = %v, want %v", gotTypeNames, tt.wantTypeNames)
			}
		})
	}
}
