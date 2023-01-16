/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"testing"
)

func Test_goID(t *testing.T) {

	var threadID ThreadID
	var err error
	var id int

	threadID = goID()
	if !threadID.IsValid() {
		t.Errorf("threadID not valid: %q", string(threadID))
	}
	s := threadID.String()
	if s == "" {
		t.Error("threadID.String empty string")
	}
	if id, err = strconv.Atoi(s); err != nil {
		t.Errorf("strconv.Atoi failed: %q", s)
	}
	if id < 1 {
		t.Errorf("unexpected id: %d exp 1…", id)
	}
}
