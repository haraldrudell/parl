/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Closer is a deferrable function that closes a channel recovering
// from panic
func Closer[T any](ch chan T, errp *error) {
	defer Recover(Annotation(), errp, nil)

	close(ch)
}
