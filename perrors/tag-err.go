/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

// TagErr prepends err’s error message with tags
//   - “perrors NewPF tag1 tag2: s cannot be empty”
//   - err2 is enrued to have a stack trace from caller of TagErr
func TagErr(err error, tags ...string) (err2 error) {
	var frames = 1 // count TagErr frame

	// ensure error has stack
	if !HasStack(err) {
		err = Stackn(err, frames)
	}

	// values to print
	var s = pruntime.NewCodeLocation(frames).PackFunc()
	if tagString := strings.Join(tags, "\x20"); tagString != "" {
		s += "\x20" + tagString
	}

	err2 = Errorf("%s: %w", s, err)
	return
}
