/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

func TestNewIterator(t *testing.T) {
	//t.Errorf("logging on")
	// for debugging, disable timeouts here
	const timeoutIsDisabled = true
	var dir = t.TempDir()
	var urwx fs.FileMode = 0700
	var dir2 = filepath.Join(dir, "dir2")
	// 10 ms
	var shortTime = 100 * time.Millisecond
	// 100 ms
	var cancelTimeout = 100 * time.Millisecond

	var ctx, cancelFunc = context.WithCancel(context.Background())
	var err error
	var value *WatchEvent
	var hasValue, receivedRecord bool
	var timer *time.Timer
	var resettingTimer *ptime.ThreadSafeTimer
	var recordCh <-chan recordR[*WatchEvent]

	var iterator iters.Iterator[*WatchEvent] = NewIterator(dir, WatchOpAll, NoIgnores, ctx)

	// watcher add is deferred, so must listen first
	// iterator should have event and not timeout
	var isNext sync.WaitGroup
	isNext.Add(1)
	recordCh = invokeNext(iterator, &isNext)
	isNext.Wait()
	timer = time.NewTimer(shortTime)
	<-timer.C

	// create a file-system watch event by writing to dir
	if err = os.Mkdir(dir2, urwx); err != nil {
		panic(err)
	}

	// timeout channel, non-nil if timeouts are enabled
	var C <-chan time.Time
	if !timeoutIsDisabled {
		// [ptime.NewThreadSafeTimer] to handle Reset correctly
		resettingTimer = ptime.NewThreadSafeTimer(shortTime)
		defer resettingTimer.Stop()
		C = resettingTimer.C
	}
	select {
	case record := <-recordCh:
		receivedRecord = true
		value = record.value
		hasValue = record.hasValue
		err = record.err
	case <-C:
		t.Errorf("iterator.Next timeout %s err %s", ptime.Duration(shortTime), perrors.Short(err))
	}

	// if timeout, consult iterator.Cancel
	if !receivedRecord {
		if !timeoutIsDisabled {
			resettingTimer.Reset(cancelTimeout)
		}
		var cancelCh = invokeCancel(iterator)
		select {
		case record := <-cancelCh:
			t.Fatalf("iterator.Cancel err %s", perrors.Short(record.err))
		case <-C:
			t.Fatalf("iterator.Cancel timeout %s", ptime.Duration(cancelTimeout))
		}
	}

	// iterator.Next should not cause error
	if err != nil {
		t.Fatalf("invokeNext err: %s", perrors.Short(err))
	}

	if !hasValue {
		t.Error("Next hasValue false")
	}
	if value == nil {
		t.Fatalf("Next value nil")
	}

	t.Logf("value: %s", value.Dump())

	cancelFunc()
}

// recordR contains the result of iterator.Next
type recordR[T any] struct {
	value    T
	hasValue bool
	err      error
}

// invokeNext returns a channel for the iterator’s Next method
func invokeNext[T any](iterator iters.Iterator[T], isNext *sync.WaitGroup) (ch <-chan recordR[T]) {
	var c = make(chan recordR[T])
	go invokeNextThread(iterator, c, isNext)
	ch = c
	return
}

// invokeNextThread if goroutine that waits for Next event
func invokeNextThread[T any](iterator iters.Iterator[T], ch chan<- recordR[T], isNext *sync.WaitGroup) {
	var record recordR[T]
	defer func() { ch <- record }()
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &record.err)

	isNext.Done()
	record.value, record.hasValue = iterator.Next()
}

// invokeNext returns a channel for the iterator’s Next method
func invokeCancel[T any](iterator iters.Iterator[T]) (ch <-chan recordR[T]) {
	var c = make(chan recordR[T])
	go invokeCancelThread(iterator, c)
	ch = c
	return
}

// invokeNextThread if goroutine that waits for Next event
func invokeCancelThread[T any](iterator iters.Iterator[T], ch chan<- recordR[T]) {
	var record recordR[T]
	defer func() { ch <- record }()
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &record.err)

	record.err = iterator.Cancel()
}
