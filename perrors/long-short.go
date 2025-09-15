/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "github.com/haraldrudell/parl/perrors/errorglue"

// LongShort picks output format
//   - no error: “OK”
//   - no panic: “error-message at runtime.gopanic:26”
//   - panic: long format with all stack traces and values
//   - associated errors are always printed
func LongShort(err error) (message string) {
	var format errorglue.CSFormat
	if err == nil {
		format = errorglue.ShortFormat
	} else if isPanic, _, _, _ := IsPanic(err); isPanic {
		format = errorglue.LongFormat
	} else {
		format = errorglue.ShortFormat
	}
	message = errorglue.ChainString(err, format)

	return
}
