/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pstrings

import (
	"testing"
)

func TestFit(t *testing.T) {
	t.Logf("a: %d", len([]rune("…")))
	type args struct {
		s     string
		width int
		pad   bool
	}
	tests := []struct {
		name   string
		args   args
		wantS2 string
	}{
		{"nil", args{"", 0, false}, ""},
		{"width 0", args{"ab", 0, false}, "ab"},
		{"length == width", args{"ab", 2, false}, "ab"},
		{"pad", args{"a", 2, true}, "a "},
		{"cut", args{"abcdef", 4, true}, "ab…f"},
		{"failure1", args{
			"/System/Library/CoreServices/Applications/Screen Sharing.app/Contents/MacOS/Screen Sharing",
			59, true},
			"/System/Library/CoreServices/A…ontents/MacOS/Screen Sharing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS2 := Fit(tt.args.s, tt.args.width, tt.args.pad); gotS2 != tt.wantS2 {
				t.Errorf("Fit() = %d %q, want %q", len(gotS2), gotS2, tt.wantS2)
			}
		})
	}
}
