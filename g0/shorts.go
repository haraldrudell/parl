/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
)

func Shorts(threads []parl.ThreadData) (s string) {
	length := len(threads)
	s = "threads:" + strconv.Itoa(length)
	if length > 0 {
		sList := make([]string, length)
		for i, t := range threads {
			sList[i] = t.Short()
		}
		s += "[" + strings.Join(sList, ",") + "]"
	}
	return
}
