/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"slices"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

// errStore is a slice to distinguish multiple invocations of AddErr
//   - must be initialization free
//   - must be thread-safe
type errStore struct {
	errsLock    sync.Mutex
	errs        []error
	count       parl.Atomic64[int]
	IsFirstLong atomic.Bool
}

func (e *errStore) Count() (count int) { return e.count.Load() }

func (e *errStore) Add(err error) {
	e.errsLock.Lock()
	defer e.errsLock.Unlock()

	e.errs = append(e.errs, err)
	e.count.Add(1)
}

func (e *errStore) GetN(index ...int) (err error) {
	e.errsLock.Lock()
	defer e.errsLock.Unlock()

	var i int
	if len(index) > 0 {
		i = index[0]
	}
	if i >= 0 && i < len(e.errs) {
		err = e.errs[i]
	}
	return
}

func (e *errStore) Get() (errs []error) {
	e.errsLock.Lock()
	defer e.errsLock.Unlock()

	errs = slices.Clone(e.errs)
	return
}
