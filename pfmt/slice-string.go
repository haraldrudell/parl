/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package pfmt

import (
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pslices"
)

// "2[rob,pike]"
func SliceString[E any](slic []E) (s string) {
	return parl.Sprintf("%d[%s]", len(slic), strings.Join(pslices.StringifySlice(slic), ","))
}
