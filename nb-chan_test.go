/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
	"golang.org/x/exp/slices"
)

// TestNBChan initialization-free [NBChan], [NBChan.Ch]
func TestNBChanNoNew(t *testing.T) {
	var ch <-chan int

	// NBChan no initialization, Ch()
	var nbChan NBChan[int]
	if ch = nbChan.Ch(); ch == nil {
		t.Errorf("Ch returned nil")
	}

	// check for channel errors
	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

// TestNBChanNew: [NewNBChan], [NBChan], [NBChan.Ch]
func TestNBChanNew(t *testing.T) {
	var nbChan NBChan[int]
	var ch <-chan int

	nbChan = *NewNBChan[int]()
	if ch = nbChan.Ch(); ch == nil {
		t.Errorf("NewNBChan Ch returned nil")
	}

	// check for channel errors
	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

// TestNBChanCount: [NBChan.Count] [NBChan.Send]
func TestNBChanCount(t *testing.T) {
	var expCount0 = 0
	var value1 = 3
	var expCount1 = 1

	var actualInt int
	var nbChan NBChan[int]

	// Count()
	if actualInt = nbChan.Count(); actualInt != expCount0 {
		t.Errorf("count0 %d exp %d", actualInt, expCount0)
	}
	nbChan.Send(value1)
	if actualInt = nbChan.Count(); actualInt != expCount1 {
		t.Errorf("count1 %d exp %d", actualInt, expCount1)
	}

	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

// TestNBChanScavenge: [NBChan.Scavenge] [NBChan.SendMany]
func TestNBChanScavenge(t *testing.T) {
	var expLength0 = 0
	var expCapacity0 = 0
	var value1 = []int{3, 4}
	var expLength1 = len(value1)
	var expCapacity1 = defaultNBChanSize
	var scavenge2 = 1 // minimum Scavenge value
	var expLength2 = len(value1)
	var expCapacity2 = len(value1) - 1 // 1 with thread, outputQueue unallocated

	var actualLength, actualCapacity int

	var nbChan NBChan[int]

	// initial Scavenge
	actualLength = nbChan.Count()
	actualCapacity = nbChan.Capacity()
	if actualLength != expLength0 {
		t.Errorf("expLength0 %d exp %d", actualLength, expLength0)
	}
	if actualCapacity != expCapacity0 {
		t.Errorf("expCapacity0 %d exp %d", actualLength, expCapacity0)
	}

	// default allocation size 10
	nbChan.SendMany(value1)
	actualLength = nbChan.Count()
	actualCapacity = nbChan.Capacity()
	if actualLength != expLength1 {
		t.Errorf("expLength1 %d exp %d", actualLength, expLength1)
	}
	if actualCapacity != expCapacity1 {
		t.Errorf("expCapacity1 %d exp %d", actualCapacity, expCapacity1)
	}
	nbChan.Scavenge(scavenge2) // scavenge to minimum size
	actualLength = nbChan.Count()
	actualCapacity = nbChan.Capacity()
	if actualLength != expLength2 {
		t.Errorf("expLength2 %d exp %d", actualLength, expLength2)
	}
	if actualCapacity != expCapacity2 {
		t.Errorf("expCapacity2 %d exp %d", actualCapacity, expCapacity2)
	}

	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

// [NBChan.SetAllocationSize]
func TestNBChanSetAllocationSize(t *testing.T) {
	var size = 100
	var value1 = []int{3, 4}
	var expLength1 = len(value1)
	var expCapacity1 = size

	var actualLength, actualCapacity int

	var nbChan NBChan[int]
	nbChan.SetAllocationSize(size).SendMany(value1)
	actualLength = nbChan.Count()
	actualCapacity = nbChan.Capacity()
	if actualLength != expLength1 {
		t.Errorf("expLength1 %d exp %d", actualLength, expLength1)
	}
	if actualCapacity != expCapacity1 {
		t.Errorf("expCapacity1 %d exp %d", actualCapacity, expCapacity1)
	}

	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

type NBChanReceiver[T any] struct {
	isReady      sync.WaitGroup
	value        T
	valueIsValid bool
	didRecive    bool
	isExit       sync.WaitGroup
}

func NewNBChanReceiver[T any]() (n *NBChanReceiver[T]) { return &NBChanReceiver[T]{} }
func (n *NBChanReceiver[T]) Receive(ch <-chan T) (n2 *NBChanReceiver[T]) {
	n2 = n
	n.isReady.Add(1)
	n.isExit.Add(1)
	go n.thread(ch)
	return
}
func (n *NBChanReceiver[T]) thread(ch <-chan T) {
	defer n.isExit.Done()

	n.isReady.Done()
	n.value, n.valueIsValid = <-ch
	n.didRecive = true
}

// [NBChan.Ch] channel-read
func TestNBChanReceive(t *testing.T) {
	var value = 3

	var nbChan NBChan[int]
	var n = NewNBChanReceiver[int]().Receive(nbChan.Ch())
	n.isReady.Wait()
	if n.didRecive {
		t.Error("empty NBChan receive")
	}
	nbChan.Send(value)
	n.isExit.Wait()
	if !n.didRecive {
		t.Error("didReceive false")
	}
	if !n.valueIsValid {
		t.Error("valueIsValid false")
	}
	if n.value != value {
		t.Errorf("received %d exp %d", n.value, value)
	}
	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

// [NBChan.Get]
func TestNBChanGet(t *testing.T) {
	var values = []int{3, 4, 5, 6}
	var getArg = 2
	var exp1 = values[:getArg]
	var exp2 = values[getArg:]

	var actual []int

	var nbChan NBChan[int]

	// Get with limit
	nbChan.SendMany(values)
	actual = nbChan.Get(getArg)
	if !slices.Equal(actual, exp1) {
		t.Errorf("Get %d: '%v' exp '%v'", getArg, actual, exp1)
	}

	// Get all
	actual = nbChan.Get()
	if !slices.Equal(actual, exp2) {
		t.Errorf("Get all: '%v' exp '%v'", actual, exp2)
	}

	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

type NBChanWaitForClose[T any] struct {
	isReady  sync.WaitGroup
	err      error
	didClose bool
	isExit   sync.WaitGroup
}

func NewNBChanWaitForClose[T any]() (n *NBChanWaitForClose[T]) { return &NBChanWaitForClose[T]{} }
func (n *NBChanWaitForClose[T]) Wait(nbChan *NBChan[T]) (n2 *NBChanWaitForClose[T]) {
	n2 = n
	n.isReady.Add(1)
	n.isExit.Add(1)
	go n.thread(nbChan)
	return
}
func (n *NBChanWaitForClose[T]) thread(nbChan *NBChan[T]) {
	defer n.isExit.Done()

	n.isReady.Done()
	nbChan.WaitForClose(&n.err)
	n.didClose = true
}

// [NBChan.Close] [NBChan.DidClose] [NBChan.IsClosed] [NBChan.WaitForClose]
func TestNBChanClose(t *testing.T) {
	var value = 3

	var didClose bool
	var timer *time.Timer
	var ok bool
	var actValue int

	var nbChan NBChan[int]

	// NBChan with value:
	//	- Close was not invoked
	//	- underlying channel is not closed
	var n = NewNBChanWaitForClose[int]().Wait(&nbChan)
	t.Log("n.isReady.Wait…")
	n.isReady.Wait()
	t.Log("n.isReady.Wait complete")
	// send a value to launch a thread
	nbChan.Send(value)
	// Close should not have been invoked
	if nbChan.DidClose() {
		t.Error("DidClose0")
	}
	// underlying channel should not be closed
	if nbChan.IsClosed() {
		t.Error("IsClosed0")
	}
	// internal close state should be false
	if n.didClose {
		t.Error("n.didClose0")
	}

	// non-empty closed NBChan
	//	- close is deferred
	//	- the channel isn’t actually closed yet
	didClose = nbChan.Close()
	// close should be deferred
	if didClose {
		t.Error("didClose1 true")
	}
	// Close should have been invoked
	if !nbChan.DidClose() {
		t.Error("DidClose1 false")
	}
	// underlying channel should not be closed
	if nbChan.IsClosed() {
		t.Error("IsClosed1")
	}
	// internal close state should be false
	if n.didClose {
		t.Error("n.didClose1")
	}

	// NBChan after deferred close:
	//	- reading the channel will cause thread exit
	//	- thread exit executes deferred close
	// empty the channel: executes deferred close
	if actValue, ok = <-nbChan.Ch(); !ok {
		t.Error("Sent item not on channel")
	}
	if actValue != value {
		t.Errorf("received %d exp %d", actValue, value)
	}
	// the channel should then close
	if _, ok = <-nbChan.Ch(); ok {
		t.Error("Channel did not close")
	}
	// Close should have been invoked
	if !nbChan.DidClose() {
		t.Error("DidClose2 false")
	}
	t.Log("n.isExit.Wait…")
	n.isExit.Wait()
	t.Log("n.isExit.Wait complete")
	// internal close state should be true
	if !n.didClose {
		t.Error("n.didClose2 false")
	}
	if n.err != nil {
		t.Errorf("WaitForClose err: %s", perrors.Short(n.err))
	}
	// underlying channel should be closed
	if !nbChan.IsClosed() {
		t.Error("IsClosed2")
	}

	// a deferred close should close the data channel
	//	- max 1 ms wait to error
	timer = time.NewTimer(time.Millisecond)
	select {
	case <-timer.C:
		t.Error("DataWaitCh not closed by deferred Close")
	case <-nbChan.DataWaitCh():
	}
	timer.Stop()

	// subsequent close should not do anything
	didClose = nbChan.Close()
	if didClose {
		t.Error("didClose3 true")
	}

	// there should be no errors
	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

// [NBChan.CloseNow]
func TestNBChanCloseNow(t *testing.T) {
	var value = 3

	var didClose bool
	var err error
	var timer *time.Timer

	// NBChan with thread in channel send:
	//	- should not be closed and
	//	- channel should not be closed
	var nbChan NBChan[int]
	nbChan.Send(value)
	// Close should not have been invoked
	if nbChan.DidClose() {
		t.Error("DidClose0")
	}
	// underlying channel should not be closed
	if nbChan.IsClosed() {
		t.Error("IsClosed0")
	}

	// after CloseNow on non-empty channel:
	//	- didClose true for first CloseNow invocation
	//	- DidClose true
	//	- channel is closed
	//	- NBChan is empty
	//	- queues are nil
	//	- thread is exited
	didClose, err = nbChan.CloseNow()
	// internal close should be true
	if !didClose {
		t.Error("didClose1 false")
	}
	// there should be no errors
	if err != nil {
		t.Errorf("CloseNow err: %s", perrors.Short(err))
	}
	// Close should have been invoked
	if !nbChan.DidClose() {
		t.Error("DidClose1 false")
	}
	// underlying channel should be closed
	if !nbChan.IsClosed() {
		t.Error("IsClosed1 false")
	}
	// NBChan should be empty
	if nbChan.Count() != 0 {
		t.Errorf("Count %d exp %d", nbChan.Count(), 0)
	}
	// inputQueue should be discarded
	if nbChan.inputQueue != nil {
		t.Error("inputQueue not nil")
	}
	// outputQueue should be discarded
	if nbChan.outputQueue != nil {
		t.Error("outputQueue not nil")
	}
	// thread should have exit
	if nbChan.isRunningThread.Load() {
		t.Error("isRunningThread true")
	}
	// thread waiter should not wait
	t.Log("threadWait.Wait…")
	nbChan.waitForSendThread()
	t.Log("threadWait.Wait complete")
	// data waiter should not wait
	//	- error in 1 ms
	timer = time.NewTimer(time.Millisecond)
	select {
	case <-timer.C:
		t.Error("DataWaitCh not closed")
	case <-nbChan.DataWaitCh():
	}
	timer.Stop()

	// subsequent CloseNow
	//	- didClose is false
	//	- no errors
	didClose, err = nbChan.CloseNow()
	if didClose {
		t.Error("didClose2 true")
	}
	if err != nil {
		t.Errorf("CloseNow2 err: %s", perrors.Short(err))
	}

	// check for any errors
	if err := nbChan.GetError(); err != nil {
		cL := pruntime.NewCodeLocation(0)
		loc := cL.FuncIdentifier() + ":" + strconv.Itoa(cL.Line)
		t.Errorf("%s nbChan.GetError: %s", loc, perrors.Short(err))
	}
}

// exit with unused NBChan
func TestNBChanExit1(t *testing.T) {
	var nbChan = *NewNBChan[int]()
	_ = &nbChan
}

// exit with thread at channel send
func TestNBChanExit2(t *testing.T) {
	var nbChan NBChan[int]
	nbChan.Send(1)
}

// exit after thread shutdown
func TestNBChanExit3(t *testing.T) {
	var nbChan NBChan[int]
	nbChan.Send(1)
	nbChan.Get()
}

// [NBChanAlways]
func TestNBChanAlways(t *testing.T) {
	var value = 3
	var value2 = 4
	var expValues = []int{value}

	var nbChan NBChan[int]
	var values []int
	var timer *time.Timer

	// create, send and empty channel
	nbChan = *NewNBChan[int](NBChanAlways)
	nbChan.Send(value)
	values = nbChan.Get()
	if !slices.Equal(values, expValues) {
		t.Errorf("Get0 '%v' exp '%v'", values, expValues)
	}
	// wait up to 1 ms for proper thread status
	timer = time.NewTimer(time.Millisecond)
	for nbChan.ThreadStatus() != NBChanAlert {
		select {
		case <-timer.C:
		default:
			continue
		}
		break
	}
	timer.Stop()
	if nbChan.ThreadStatus() != NBChanAlert {
		t.Errorf("ThreadStatus %s exp %s", nbChan.ThreadStatus(), NBChanAlert)
	}
	nbChan.Send(value2)
	timer = time.NewTimer(time.Millisecond)
	for nbChan.ThreadStatus() != NBChanSendBlock {
		select {
		case <-timer.C:
		default:
			continue
		}
		break
	}
	if nbChan.ThreadStatus() != NBChanSendBlock {
		t.Errorf("ThreadStatus %s exp %s", nbChan.ThreadStatus(), NBChanSendBlock)
	}
	nbChan.CloseNow()
}

type ChWaiter struct {
	isReady, isClosed sync.WaitGroup
	didClose          atomic.Bool
}

func NewChWaiter() (c *ChWaiter) {
	return &ChWaiter{}
}
func (c *ChWaiter) Wait(ch <-chan struct{}) (c2 *ChWaiter) {
	c2 = c
	c.isReady.Add(1)
	c.isClosed.Add(1)
	go c.WaitThread(ch)
	return
}
func (c *ChWaiter) WaitThread(ch <-chan struct{}) {
	defer c.isClosed.Done()
	defer c.didClose.Store(true)

	c.isReady.Done()
	<-ch
}

// [NBChanNone] [NBChan.DataWaitCh]
func TestNBChanNone(t *testing.T) {
	var value = 3

	var nbChan NBChan[int]
	var waiter *ChWaiter
	//var timer *time.Timer

	// empty channel should causes wait
	nbChan = *NewNBChan[int](NBChanNone)
	waiter = NewChWaiter().Wait(nbChan.DataWaitCh())
	// await ChWaiter thread
	waiter.isReady.Wait()
	// ensure ChWaiter is waiting
	if waiter.didClose.Load() {
		t.Fatal("DataWaitCh0 closed")
	}

	// closed empty channel should not wait
	nbChan.Close()
	if waiter.didClose.Load() {
		t.Fatal("DataWaitCh0 still closed")
	}

	// channel with item should not wait
	nbChan = *NewNBChan[int](NBChanNone)
	nbChan.Send(value)
	waiter = NewChWaiter().Wait(nbChan.DataWaitCh())
	// await ChWaiter thread
	waiter.isReady.Wait()
	if !waiter.didClose.Load() {
		t.Fatal("DataWaitCh1 not closed")
	}

	// channel becoming empty should wait
	nbChan = *NewNBChan[int](NBChanNone)
	nbChan.Send(value)
	nbChan.Get()
	waiter = NewChWaiter().Wait(nbChan.DataWaitCh())
	// await ChWaiter thread
	waiter.isReady.Wait()
	if waiter.didClose.Load() {
		t.Fatal("DataWaitCh2 closed")
	}
}
