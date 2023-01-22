/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "sync"

type Queue[T any] struct {
	lock  sync.RWMutex
	q0, q []T
}

func NewQueue[T any]() (queue *Queue[T]) {
	return &Queue[T]{}
}

func (q *Queue[T]) Add(value T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.q = append(q.q, value)
}

func (q *Queue[T]) AddSlice(slice []T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.q = append(q.q, slice...)
}

func (q *Queue[T]) Remove(value T, ok bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	length := len(q.q)
	if length > 0 {
		value = q.q[0]
		if length > 1 {
			//copy()
			_ = length
		}
	}
	_ = value
	_ = q.q0
}
