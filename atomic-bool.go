/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync/atomic"

/*
AtomicBool is a thread-safe flag.
AtomicBool requires no initialization
 var isDone parl.AtomicBool
 if isDone.Set() // isDone was not set, but is set now
 …
 if !isDone.IsTrue() // isDone is not set
*/
type AtomicBool struct {
	value int32 // atomic
}

const (
	abTrue  = int32(1)
	abFalse = int32(0)
)

func (ab *AtomicBool) IsTrue() (isTrue bool) {
	return atomic.LoadInt32(&ab.value) == abTrue
}

func (ab *AtomicBool) Set() (wasNotSet bool) {
	return atomic.SwapInt32(&ab.value, abTrue) != abTrue
}

func (ab *AtomicBool) Clear() (wasSet bool) {
	return atomic.SwapInt32(&ab.value, abFalse) == abTrue
}
