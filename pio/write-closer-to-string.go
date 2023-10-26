/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// On write after close, ErrFileAlreadyClosed is returned.
//
//	if errors.Is(err, pio.ErrFileAlreadyClosed)…
var ErrFileAlreadyClosed = errors.New("file alread closed")

// WriteCloserToString is an io.WriteCloser that aggregates its oputput in a string. Thread-safe.
//   - the string is available using the Data method.
type WriteCloserToString struct {
	isClosed atomic.Bool
	lock     sync.Mutex
	s        string
}

// NewWriteCloserToString returns an io.WriteCloser that aggregates its oputput in a string. Thread-safe.
func NewWriteCloserToString() io.WriteCloser {
	return &WriteCloserToString{}
}

// Write always succeeds
func (wc *WriteCloserToString) Write(p []byte) (n int, err error) {
	if wc.isClosed.Load() {
		err = perrors.ErrorfPF(ErrFileAlreadyClosed.Error())
		return
	}

	wc.lock.Lock()
	defer wc.lock.Unlock()

	wc.s += string(p)
	n = len(p)
	return
}

// Close should only be invoked once.
// Close is not required for releasing resources.
func (wc *WriteCloserToString) Close() (err error) {
	wc.isClosed.Store(true)
	return
}

// Data returns current string data
func (wc *WriteCloserToString) Data() (s string) {
	wc.lock.Lock()
	defer wc.lock.Unlock()

	return wc.s
}
