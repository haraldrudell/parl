/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
)

type FanOut struct {
	wg      sync.WaitGroup
	ErrCh   chan error
	Results chan interface{}
}

type FanThunk func() (result interface{}, err error)
type FanProc func() (err error)

func NewFanOut() (fo *FanOut) {
	return &FanOut{
		ErrCh:   make(chan error),
		Results: make(chan interface{}),
	}
}

// Do executes a procedure in a goroutine that has no result other than a possible non-nil error
func (cr *FanOut) Do(name string, proc FanProc) {
	cr.wg.Add(1)
	go cr.fanout(name, nil, proc)
}

// Run executes a thunk in a goroutine with a possible non-nil result and a possible non-nil error
func (cr *FanOut) Run(name string, thunk FanThunk) {
	cr.wg.Add(1)
	go cr.fanout(name, thunk, nil)
}

func (cr *FanOut) fanout(name string, thunk FanThunk, proc FanProc) {
	defer cr.wg.Done()
	var err error
	defer Recover(name, &err, func(e error) { cr.ErrCh <- err })

	if thunk == nil {
		err = proc()
	} else {
		var result interface{}
		result, err = thunk()
		if result != nil {
			cr.Results <- result
		}
	}
}

// Wait waits for all Do and Run invocations to complete, then shuts down
func (cr *FanOut) Wait() {
	cr.wg.Wait()
	close(cr.ErrCh)
	close(cr.Results)
}
