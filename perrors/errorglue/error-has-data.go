/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// ErrorHasData enrichens an error with key and value strings
type ErrorHasData interface {
	KeyValue() (key, value string)
}
