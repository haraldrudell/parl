/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker2"
	"github.com/haraldrudell/parl/perrors/errorglue"
)

// Deferr prints error message or “OK” on function return
//   - label: message prepended. A colon and a space is appended to label.
//   - errp: pointer to error
//   - — if *errp contains an error it is printed in [Short] form
//   - — “label: Error message at error116.(*csTypeName).FuncName-chainstring_test.go:26”
//   - — if *errp is nil, “OK” is printed
//   - — errp nil is panic
//   - fn: optional printf function eg. parl.Log. If nil nothing is printed
//   - —
//   - deferrable and returns the message
//
// Usage:
//
//	func f() (err error) {
//	  defer perrors.Deferr("bad", &err, parl.Log)
//	  defer func() { fmt.Println(Deferr("bad", &err)) }()
func Deferr(label string, errp *error, fn func(format string, a ...any)) (message string) {

	// errp nil case
	cyclebreaker2.NilPanic("errp", errp)

	if label != "" {
		label += ":\x20"
	}
	message = label + errorglue.ChainString(*errp, errorglue.ShortFormat)
	if fn != nil {
		fn(message)
	}

	return
}
