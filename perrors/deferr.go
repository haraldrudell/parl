/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"github.com/haraldrudell/parl/perrors/errorglue"
)

// Deferr invokes a printing function for an error pointer
//   - label: message prepended. A colon and a space is appended to label.
//   - errp: pointer to error
//   - — if *errp contains an error it is printed in [Short] form
//   - — “label: Error message at error116.(*csTypeName).FuncName-chainstring_test.go:26”
//   - — if *errp is nil, “OK” is printed
//   - errp nil: message returned bnut not printed: “perrors.Deferr: errp nil”
//   - fn: optional printf function eg. parl.Log. If missing nothing is printed
//   - —
//   - deferrable and returns the message
func Deferr(label string, errp *error, fn func(format string, a ...interface{})) (message string) {
	if errp == nil {
		return "perrors.Deferr: errp nil"
	}
	if label != "" {
		label += ":\x20"
	}
	s := label + errorglue.ChainString(*errp, errorglue.ShortFormat)
	if fn != nil {
		fn(s)
	}

	return s
}
