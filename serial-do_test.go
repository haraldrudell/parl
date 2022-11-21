/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

func TestSerialDo(t *testing.T) {
	ctx := NewCancelContext(context.Background())

	var lock sync.Mutex
	var events []*SerialDoEvent
	var busySince time.Time
	var pendingSince time.Time
	var isCancel bool
	var isWaitComplete bool
	var thunkWait sync.WaitGroup
	thunkWait.Add(1)
	var thunks int64 // atomic

	// create serialDo
	serialDo := NewSerialDo(func(at time.Time) {
		atomic.AddInt64(&thunks, 1)
		thunkWait.Wait()
	}, func(event *SerialDoEvent) {
		lock.Lock()
		defer lock.Unlock()
		events = append(events, event)
	}, func(err error) {
		t.Errorf("serialDo err: " + perrors.Long(err))
		t.FailNow()
	}, ctx)
	defer serialDo.Wait()
	defer InvokeCancel(ctx)

	busySince, pendingSince, isCancel, _ = serialDo.State()
	if !busySince.IsZero() || !pendingSince.IsZero() || isCancel {
		t.Errorf("Bad initial state: " + serialDo.String())
	}

	serialDo.Do() // busy
	serialDo.Do() // pending

	busySince, pendingSince, isCancel, _ = serialDo.State()
	if busySince.IsZero() || pendingSince.IsZero() || isCancel {
		t.Errorf("Bad initial state: " + serialDo.String())
	}

	// cancel serialDo
	InvokeCancel(ctx)
	// allow thunk to complete
	thunkWait.Done()
	// wait for serialDo to complete
	serialDo.Wait()

	thunkCount := int(atomic.LoadInt64(&thunks))
	if thunkCount != 1 {
		t.Errorf("FAIL thunkCOunt %d exp 1", thunkCount)
	}
	busySince, pendingSince, isCancel, isWaitComplete = serialDo.State()
	if !busySince.IsZero() || !pendingSince.IsZero() || !isCancel || !isWaitComplete {
		t.Errorf("Bad initial state: " + serialDo.String())
	}
}
