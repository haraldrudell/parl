/*
© 2022–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "errors"

// ErrEndCallbacks indicates upon retun from a callback function that
// no more callbacks are desired. It does not indicate an error and is not returned
// as an error by any other funciton that the callback.
//
//	if errors.Is(err, parl.ErrEndCallbacks) { …
var ErrEndCallbacks = errors.New("end callbacks error")
