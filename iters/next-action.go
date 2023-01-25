/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import "strconv"

const (
	IsSame NextAction = 0 // IsSame indicates to Delegate.Next that this is a Same-type incovation
	IsNext NextAction = 1 // IsNext indicates to Delegate.Next that this is a Next-type incovation
)

// NextAction is a unique named type that indicates whether
// the next or the same value again is sought by Delegate.Next
type NextAction uint8

func (na NextAction) String() (s string) {
	var ok bool
	if s, ok = nextActionSet[na]; ok {
		return
	}
	s = "?\x27" + strconv.Itoa(int(na)) + "\x27"
	return
}

// nextActionSet is the set helper for NextAction
var nextActionSet = map[NextAction]string{
	IsSame: "IsSame",
	IsNext: "IsNext",
}
