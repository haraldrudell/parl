/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync/atomic"

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
