/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pstrings

import (
	"time"

	"github.com/haraldrudell/parl/perrors"
)

func IsDefaultValue(pt interface{}) (isDefault bool) {
	switch p := pt.(type) {
	case *bool:
		return !*p
	case *time.Duration:
		return *p == 0
	case *float64:
		return *p == 0
	case *int64:
		return *p == 0
	case *int:
		return *p == 0
	case *string:
		return *p == ""
	case *uint64:
		return *p == 0
	case *uint:
		return *p == 0
	case *[]string:
		return len(*p) == 0
	default:
		panic(perrors.Errorf("unknown pointer type: %T", p))
	}
}
