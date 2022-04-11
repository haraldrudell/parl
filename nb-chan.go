/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

/*
NBChan is a non-blocking thread-safe channel.
NBChan does not need initialization.
NBChan can be used for error channels.
*/
type NBChan[T any] struct {
	closerLock sync.RWMutex
	closer     ClosableChan[T] // inside lock

	// GetError()
	perrors.ParlError // possible thread panic

	sqLock      sync.Mutex
	unsentCount int  // inside lock
	sendQueue   []T  // inside lock
	hasThread   bool // inside lock
}

// NewNBChan instantiates a non-blocking channel based on existing channel.
// NewNBChan otherwise does not need initialization and can be used like:
//  var nbChan NBChan[error]
//  go wherever(nbChan.Ch())
// A non-blocking channel is a trillion-buffer channel that does not block if a reader
// is not present
func NewNBChan[T any](ch chan T) (nbChan *NBChan[T]) {
	var nb NBChan[T] // somewhere to store ch
	nb.getCh(ch)     // initialize nb based on ch
	return &nb
}

// Ch obtains the channel
func (nb *NBChan[T]) Ch() (ch chan T) {
	return nb.getCh(nil)
}

// Send sends non-blocking on the channel
func (nb *NBChan[T]) Send(value T) {
	nb.sqLock.Lock()
	defer nb.sqLock.Unlock()

	nb.unsentCount++

	// launch thread
	if !nb.hasThread {
		nb.hasThread = true
		go nb.sendThread(value) // send err in new thread
		return
	}

	// put in queue
	nb.sendQueue = append(nb.sendQueue, value) // put err in send queue
}

// Count returns number of unsent values
func (nb *NBChan[T]) Count() (unsentCount int) {
	nb.sqLock.Lock()
	defer nb.sqLock.Unlock()

	return nb.unsentCount
}

// IsClosed indicates whether the Close method has been invoked
func (nb *NBChan[T]) IsClosed() (isClosed bool) {
	nb.getCh(nil)
	return nb.closer.IsClosed()
}

// Close ensures the channel is closed.
// Close does not panic.
// Close is thread-safe.
// Close does not return until the channel is closed.
// Upon return, all invocations have a possible close error in err.
// if errp is non-nil, it is updated with error status
func (nb *NBChan[T]) Close(errp ...*error) (err error, didClose bool) {
	nb.getCh(nil)
	return nb.closer.Close(errp...)
}

func (nb *NBChan[T]) getCh(ch0 chan T) (ch chan T) {
	nb.closerLock.Lock()
	defer nb.closerLock.Unlock()

	if nb.closer.ch == nil { // instantiate closer
		if ch0 != nil {
			ch = ch0 // base on ch0
		} else {
			ch = make(chan T) // base on new unbuffered channel
		}
		nb.closer = ClosableChan[T]{ch: ch}
	} else {
		ch = nb.closer.Ch() // get ch from Closer
	}
	return
}

func (nb *NBChan[T]) sendThread(value T) {
	defer Recover(Annotation(), nil, func(err error) {
		if pruntime.IsSendOnClosedChannel(err) {
			return // ignore if the channel was or became closed
		}
		nb.AddError(err)
	})

	ch := nb.getCh(nil)
	for {
		ch <- value // may block and panic

		var ok bool
		if value, ok = nb.getValue(); !ok {
			break
		}
	}
}

func (nb *NBChan[T]) getValue() (value T, ok bool) {
	nb.sqLock.Lock()
	defer nb.sqLock.Unlock()

	nb.unsentCount--

	// no more values: end thread
	if len(nb.sendQueue) == 0 {
		nb.hasThread = false
		return
	}

	// send next value in queue
	value = nb.sendQueue[0]
	ok = true
	copy(nb.sendQueue[0:], nb.sendQueue[1:])
	nb.sendQueue = nb.sendQueue[:len(nb.sendQueue)-1]
	return
}
