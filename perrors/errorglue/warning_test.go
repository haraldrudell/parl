/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"strings"
	"testing"
)

func TestNewWarning(t *testing.T) {
	message := "message"
	w := "warning:"

	warning := NewWarning(errors.New(message))
	if !strings.Contains(warning.Error(), w) {
		t.Errorf("Warning.Error %q missing %q", warning.Error(), w)
	}
	if !strings.Contains(warning.Error(), message) {
		t.Errorf("Warning.Error %q missing %q", warning.Error(), message)
	}
	if _, ok := warning.(*WarningType); !ok {
		t.Error("Warning not WarningType")
	}
}
