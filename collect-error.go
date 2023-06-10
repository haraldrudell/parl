/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/perrors"

// CollectError is a deferrable function that reads a single error value from errCh
//   - if a non-nil error value is received, it is appended to errp
//   - CollectError is used to wait for a goroutine sending its result on a channel
func CollectError(errCh <-chan error, errp *error) (err error) {
	var ok bool
	if err, ok = <-errCh; !ok {
		return
	} else if err != nil {
		*errp = perrors.AppendError(*errp, err)
	}
	return
}
