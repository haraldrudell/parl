/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package threadprof

import (
	"fmt"
	"sync"
	"time"
)

const (
	stInitial = "no status"
	stTime    = 10 * time.Second
)

type StatuserOn struct {
	lock   sync.Mutex
	status string
	t      time.Timer
	can    chan struct{}
	shLock sync.Once
}

func newStatuser(d time.Duration) (st *StatuserOn) {
	if d == 0 {
		d = stTime
	}
	s := StatuserOn{
		status: stInitial,
		t:      *time.NewTimer(d),
		can:    make(chan struct{}),
	}
	go s.trigThread()
	return &s
}

func (st *StatuserOn) Set(status string) (statuser Statuser) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.status = status
	return st
}

func (st *StatuserOn) Shutdown() {
	st.shLock.Do(func() {
		close(st.can)
	})
}

func (st *StatuserOn) trigThread() {
	for {
		select {
		case <-st.t.C:
			fmt.Printf("### %s", st.getStatus())
		case <-st.can:
			st.t.Stop()
		}
	}
}

func (st *StatuserOn) getStatus() (s string) {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.status
}
