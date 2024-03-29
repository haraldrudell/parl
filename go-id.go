/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/pruntime/pruntimelib"
)

func goID() (threadID ThreadID) {
	var ID, _, err = pruntimelib.ParseFirstLine(pruntime.FirstStackLine())
	if perrors.IsPF(&err, "pruntime.ParseFirstLine %w", err) {
		panic(err)
	}
	threadID = ThreadID(ID)

	return
}
