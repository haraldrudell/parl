/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"io"
	"sync"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

// Closers collects io.Closer objects so they can be closed all at once.
// Closer is required for servers that may have the server itself and
type Closers struct {
	lock      sync.Mutex
	closers   []io.Closer
	closerMap map[io.Closer]int
}

func (cl *Closers) Add(closer io.Closer) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if cl.closerMap == nil {
		cl.closerMap = map[io.Closer]int{}
	} else if _, ok := cl.closerMap[closer]; ok {
		panic(perrors.NewPF("duplicate closer"))
	}

	cl.closerMap[closer] = len(cl.closers)
	cl.closers = append(cl.closers, closer)
}

func (cl *Closers) Remove(closer io.Closer) (ok bool) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if cl.closerMap == nil {
		return
	}
	var index int
	if index, ok = cl.closerMap[closer]; !ok {
		return
	}

	delete(cl.closerMap, closer)
	cl.closers = slices.Delete(cl.closers, index, index+1)

	return
}

func (cl *Closers) EnsureClosed(closer io.Closer) (err error) {
	if !cl.Remove(closer) {
		return // was not imn map, probably already closed
	}
	Close(closer, &err)
	return
}

func (cl *Closers) Close() (err error) {
	for _, c := range cl.getClosers() {
		Close(c, &err)
	}
	return
}

func (cl *Closers) getClosers() (closers []io.Closer) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	closers = cl.closers
	cl.closers = nil
	cl.closerMap = nil
	return
}
