/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0debug

import (
	"testing"

	"github.com/haraldrudell/parl"
)

func TestGoErrorDump(t *testing.T) {
	type args struct {
		goError parl.GoError
	}
	tests := []struct {
		name  string
		args  args
		wantS string
	}{
		{"nil", args{nil}, "parl.GoError: type: <nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := GoErrorDump(tt.args.goError); gotS != tt.wantS {
				t.Errorf("GoErrorDump() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
	//t.Fail()
}
