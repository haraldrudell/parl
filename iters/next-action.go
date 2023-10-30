/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import "fmt"

func (a NextAction) String() (s string) {
	var ok bool
	if s, ok = nextActionSet[a]; ok {
		return
	}
	s = fmt.Sprintf("?“%d”", a)
	return
}

// nextActionSet is the set helper for NextAction
var nextActionSet = map[NextAction]string{
	IsSame: "IsSame",
	IsNext: "IsNext",
}
