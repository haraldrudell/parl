/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package pfmt

import (
	"fmt"
	"strings"
)

type structWithPrivateFieldAny struct {
	any
}

// NoRecurseVPrint returns the reflection string representation of value
// without invoking the String method.
func NoRecurseVPrint(value any) (s string) {
	s = fmt.Sprint(structWithPrivateFieldAny{any: value})
	s = strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	return
}
