/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import (
	"os"

	"github.com/haraldrudell/parl/plog"
)

var output = func() (output func(calldepth int, s string) error) {
	return plog.GetLog(os.Stderr).Output
}()

func Log(format string, a ...any) {
	if err := output(0, Sprintf(format, a...)); err != nil {
		panic(err)
	}
}
