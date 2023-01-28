/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import "context"

type Go interface {
	AddError(err error)
	Done(err *error)
	Context() (ctx context.Context)
}
