/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"
	"testing"

	"github.com/haraldrudell/parl/pstrings"
)

func TestElemsX(t *testing.T) {
	expDirs := []string{"x", "y"}
	expFile := "z.txt"
	input := filepath.Join(filepath.Join(expDirs...), expFile)
	t.Logf("input: %q", input)

	var dirs []string
	var file string

	dirs, file = Elems(input)

	if file != expFile {
		t.Errorf("file bad: %q exp %q", file, expFile)
	}

	same := len(dirs) == len(expDirs)
	if same {
		for i, dir := range expDirs {
			if dir != expDirs[i] {
				same = false
				break
			}
		}
	}
	if !same {
		t.Errorf("dirs wrong: %s exp: %s", pstrings.QuoteList(dirs), pstrings.QuoteList(expDirs))
	}
}
