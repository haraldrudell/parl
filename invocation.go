/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"time"
)

// a timed invocation created by [InvokeTimer.Invocation]
//   - holds invocation instance-data for [InvocationTimer.invocationEnd]
type Invocation[T any] struct {
	Prev, Next    atomic.Pointer[Invocation[T]]
	ThreadID      ThreadID
	Value         T
	t0            time.Time
	invocationEnd func(invocation *Invocation[T], duration time.Duration)
}

// NewInvocation adds a new invocation to [InvokeTimer]
//   - holds invocation instance-data for [InvocationTimer.invocationEnd]
func NewInvocation[T any](invocationEnd func(invocation *Invocation[T], duration time.Duration), value T) (invocation *Invocation[T]) {
	return &Invocation[T]{
		ThreadID:      goID(),
		Value:         value,
		t0:            time.Now(),
		invocationEnd: invocationEnd,
	}
}

// DeferFunc ends an invocation
//   - provides invocation instance-data for [InvocationTimer.invocationEnd]
func (i *Invocation[T]) DeferFunc() { i.invocationEnd(i, i.Age()) }

// Age returns the current age of this invocation
func (i *Invocation[T]) Age() (age time.Duration) {
	var t1 = time.Now()
	age = t1.Sub(i.t0)
	return
}
