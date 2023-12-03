/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

const (
	// IsSame indicates to Delegate.Next that
	// this is a Same-type incovation
	IsSame NextAction = false
	// IsNext indicates to Delegate.Next that
	// this is a Next-type incovation
	IsNext NextAction = true
)

// NextAction is a unique named type that indicates whether
// the next or the same value again is sought by Delegate.Next
//   - IsSame IsNext
type NextAction bool

func (a NextAction) String() (s string) { return nextActionSet[a] }

// nextActionSet is the set helper for NextAction
var nextActionSet = map[NextAction]string{
	IsSame: "IsSame",
	IsNext: "IsNext",
}
