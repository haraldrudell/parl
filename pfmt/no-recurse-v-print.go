/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfmt

import (
	"fmt"
	"strings"
)

type structWithPrivateFieldAny struct{ any }

// NoRecurseVPrint returns the reflection string representation of value
// without invoking the String method.
func NoRecurseVPrint(value any) (s string) {
	s = fmt.Sprint(structWithPrivateFieldAny{any: value})
	s = strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	return
}
