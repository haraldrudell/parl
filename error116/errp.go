/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

// error116.Errp returns a function that updates an an error pointer value
// with additional associated errors on subsequent invocations.
// It is intended to be used with parl.Recover().
// for a thread-safe version, use error116.ParlError
func Errp(errp *error) func(e error) {
	if errp == nil {
		panic("Errp with nil argument")
	}
	return func(e error) {
		*errp = AppendError(*errp, e)
	}
}
