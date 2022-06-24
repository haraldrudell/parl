/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"reflect"
	"testing"

	"github.com/haraldrudell/parl/pstrings"
)

func TestElems(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name     string
		args     args
		wantDirs []string
		wantFile string
	}{
		{"empty string", args{""}, nil, ""},
		{"just file", args{"file.ext"}, nil, "file.ext"},
		{"absolute", args{"/x/file.ext"}, []string{"/", "x"}, "file.ext"},
		{"relative", args{"x/file.ext"}, []string{"x"}, "file.ext"},
		{"no file", args{"//"}, []string{"/", ""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDirs, gotFile := Elems(tt.args.path)
			if !reflect.DeepEqual(gotDirs, tt.wantDirs) {
				t.Errorf("Elems() gotDirs = [%s], want [%s]", pstrings.QuoteList(gotDirs), pstrings.QuoteList(tt.wantDirs))
			}
			if gotFile != tt.wantFile {
				t.Errorf("Elems() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
		})
	}
}
