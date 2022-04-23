/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
)

type ConduitDo[T any] struct {
	cond      *sync.Cond
	values    []T // behind cond lock
	waitCount int // behind cond lock
	isCancel  AtomicBool
}

func NewConduit[T any](done Doneable, ctx context.Context) (conduit Conduit[T]) {
	done = EnsureDoneable(done)
	defer done.Done()

	c := ConduitDo[T]{
		cond: sync.NewCond(&sync.Mutex{}),
	}
	OnCancel(c.onCancel, ctx)
	return &c
}

func (ct *ConduitDo[T]) Put(value T) (IsCanceled bool) {
	ct.cond.L.Lock()
	defer ct.cond.L.Unlock()

	ct.values = append(ct.values, value)
	ct.cond.Signal()
	return ct.isCancel.IsTrue()
}

func (ct *ConduitDo[T]) PutSlice(values []T) (IsCanceled bool) {
	ct.cond.L.Lock()
	defer ct.cond.L.Unlock()

	ct.values = append(ct.values, values...)
	ct.cond.Signal()
	return ct.isCancel.IsTrue()
}

func (ct *ConduitDo[T]) Get() (value T, ok bool) {
	ct.cond.L.Lock()
	defer ct.cond.L.Unlock()
	isWaiting := false
	defer func() {
		if isWaiting {
			ct.waitCount--
		}
	}()

	for {
		if len(ct.values) > 0 {
			value = ct.values[0]
			ct.values = ct.values[1:]
			ok = true
			return // got value return
		}

		if ct.isCancel.IsTrue() {
			return // channel closed and empty return: zero-value, ok == false
		}

		if !isWaiting {
			isWaiting = true
			ct.waitCount++
		}
		ct.cond.Wait()
	}
}

func (ct *ConduitDo[T]) GetSlice(max int) (values []T, ok bool) {
	ct.cond.L.Lock()
	defer ct.cond.L.Unlock()
	isWaiting := false
	defer func() {
		if isWaiting {
			ct.waitCount--
		}
	}()

	for {
		if len(ct.values) > 0 {
			if max == 0 || max >= len(ct.values) { // entire ct.values
				values = ct.values
				ct.values = nil
			} else { // part of ct.values
				values = make([]T, max)
				copy(values, ct.values)
				ct.values = ct.values[max:]
			}
			ok = true
			return // got data return, ok == true
		}

		if ct.isCancel.IsTrue() {
			return // channel closed and empty return: zero-value, ok == false
		}

		if !isWaiting {
			isWaiting = true
			ct.waitCount++
		}
		ct.cond.Wait()
	}
}

func (ct *ConduitDo[T]) IsCanceled() (IsCanceled bool) {
	return ct.isCancel.IsTrue()
}

func (ct *ConduitDo[T]) IsEmpty() (isEmpty bool) {
	return ct.Count() == 0
}

func (ct *ConduitDo[T]) Count() (count int) {
	ct.cond.L.Lock()
	defer ct.cond.L.Unlock()

	return len(ct.values)
}

func (ct *ConduitDo[T]) WaitCount() (waitCount int) {
	ct.cond.L.Lock()
	defer ct.cond.L.Unlock()

	return ct.waitCount
}

func (ct *ConduitDo[T]) onCancel() {
	if ct.isCancel.Set() {
		ct.cond.Broadcast()
	}
}
