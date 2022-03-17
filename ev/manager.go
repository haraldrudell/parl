/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ev

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/error116"
)

// ThreadManager provides thread-safe management of go routines
type ThreadManager struct {
	ctx     context.Context    // context with cancel
	cancel  context.CancelFunc // cancel for exactly this context
	events  chan Event         // channel for goroutine output consumed by caller
	goIDMap sync.Map           //[GoID]bool thread-safe
	count   int64              // atomic. Because ranging goIDMap needs lock, store count outside
	Launch  time.Time
}

var _ Manager = &ThreadManager{}

var nextName int64 // atomic

// NewManager obtains instance for executing goroutines
func NewManager(ctx0 context.Context) (mgr Manager) {
	if ctx0 == nil {
		ctx0 = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx0)
	return &ThreadManager{ctx: ctx, cancel: cancel, events: make(chan Event)}
}

// Events gets the channel emitting messages from goroutines
func (mgr *ThreadManager) Events() (out EventRx) {
	return mgr.events
}

// CalleeContext produces an environment for calling a goroutine - thread safe
func (mgr *ThreadManager) CalleeContext(ID ...string) (env Callee) {
	threadNo := atomic.AddInt64(&nextName, 1)
	var name string
	if len(ID) > 0 {
		name = ID[0]
	}
	if name == "" {
		name = fmt.Sprintf("thread-%d", threadNo)
	}
	thread := NewEvThread(name)
	env = NewCallee(name, thread.ID, mgr.events, mgr.ctx)
	mgr.goIDMap.Store(thread.ID, thread)
	atomic.AddInt64(&mgr.count, 1)
	return
}

// ProcessExit indicates whether this event is the final event from a terminating goroutine
func (mgr *ThreadManager) ProcessEvent(ev Event) (err error) {
	if ev == nil {
		return error116.New("goroutine channel closed or goroutine sent nil")
	}
	gID := ev.GoID()
	if _, ok := mgr.goIDMap.Load(gID); !ok {
		return error116.New("event from unknown goroutine")
	}
	if _, ok := ev.(*ExitEvent); ok {
		mgr.goIDMap.Delete(gID)
		atomic.AddInt64(&mgr.count, -1) // decrement
	}
	return
}

// Action determines whether the executable should continue to wait for additional threads
func (mgr *ThreadManager) Action(threadResult error, action CancelAction) (isEnd bool) {
	// a goroutine terminated with threadResult
	if action == TerminateAll || // on first terminating thread, all threads are terminated
		(action == WhileOk && threadResult != nil) { // go until error, and this is failing thread
		return mgr.Cancel()
	}
	// state: KeepGoing or (WhileOk and no error)
	if threadResult != nil { // ev.Keepgoing and errors
		strs := []string{threadResult.Error()}
		if errorList, ok := threadResult.(error116.ErrorHasList); ok {
			for _, e := range errorList.ErrorList() {
				strs = append(strs, e.Error())
			}
		}
		parl.Info("goroutine errors: %v", strings.Join(strs, "\n"))
	}
	return mgr.IsEnd()
}

// IsEnd determines if all goroutines have terminated
func (mgr *ThreadManager) IsEnd() (isEnd bool) {
	return mgr.Count() == 0
}

// Count determines remaining goroutines
func (mgr *ThreadManager) Count() (count int64) {
	count = atomic.LoadInt64(&mgr.count)
	return
}

func (mgr *ThreadManager) Threads() (names []string, IDs []GoID) {
	mgr.goIDMap.Range(func(key, value interface{}) bool {
		thread := getMapValue(value)
		names = append(names, thread.Name)
		IDs = append(IDs, thread.ID)
		return true
	})
	return
}

func getMapValue(value interface{}) (thread *EvThread) {
	var ok bool
	if thread, ok = value.(*EvThread); !ok {
		panic(parl.Errorf("Bad value type in thread map: %T", value))
	}
	return
}

// Cancel shuts down all goroutines
func (mgr *ThreadManager) Cancel() (isEnd bool) {
	mgr.cancel()
	return mgr.IsEnd()
}
