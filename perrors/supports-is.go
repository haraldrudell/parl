/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

// SupportsIs is used for type assertions determining if an error value
// implements the Is() method, therefore supports errors.Is()
type SupportsIs interface {
	error
	Is(target error) (isThisType bool)
}
