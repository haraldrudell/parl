/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pdebug

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime/pruntimelib"
)

// getID obtains gorutine ID, as of go1.18 a numeric string "1"…
func ParseFirstLine(debugStack []byte) (ID parl.ThreadID, status parl.ThreadStatus, err error) {

	var uID uint64
	var status0 string
	if uID, status0, err = pruntimelib.ParseFirstLine(debugStack); err != nil {
		err = perrors.Stack(err)
		return
	}
	ID = parl.ThreadID(uID)
	status = parl.ThreadStatus(status0)

	return
}
