/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type ThreadResult struct {
	Err // Error() string
}

// Err is a public error
type Err interface {
	error
}

func NewThreadResult(err error) (failure error) {
	return &ThreadResult{Err: err}
}
