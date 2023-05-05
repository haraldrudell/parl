/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "testing"

func TestIsNil(t *testing.T) {
	// t.Logf("is nil: %t", (any)((*int)(nil)) == nil)
	// t.Logf("IsNil: %t", IsNil((*int)(nil)))
	// t.Fail()

	var p *int
	var i int
	type args struct {
		v any
	}
	tests := []struct {
		name      string
		args      args
		wantIsNil bool
	}{
		{"one", args{1}, false},
		{"zero", args{0}, false},
		{"false", args{false}, false},
		{"nil", args{nil}, true},
		{"typed nil", args{p}, true},
		{"typed non-nil", args{&i}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsNil := IsNil(tt.args.v); gotIsNil != tt.wantIsNil {
				t.Errorf("IsNil() = %v, want %v", gotIsNil, tt.wantIsNil)
			}
		})
	}
}
