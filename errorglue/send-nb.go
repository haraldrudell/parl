/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"sync"

	"github.com/haraldrudell/parl/runt"
)

// SendNb implements non-blocking send using a thread and buffer up to size of int
type SendNb struct {
	SendChannel
	sqLock    sync.Mutex
	errQueue  []error // inside lock
	hasThread bool    // inside lock
}

// Send sends an error on the error channel. non-blocking. Thread-safe
func (sc *SendNb) Send(err error, wgs ...*sync.WaitGroup) {
	if err == nil {
		return // nothing to send
	}
	var wg *sync.WaitGroup
	if len(wgs) > 0 {
		wg = wgs[0]
	}
	if sc.getErrCh() == nil { // thread-safe determination
		if wg != nil {
			wg.Done()
		}
		return // not initialized to send errors
	}

	sc.saveOrLaunch(err, wg) // put in slice, launch thread if necessary
}

func (sc *SendNb) saveOrLaunch(err error, wg *sync.WaitGroup) {
	sc.sqLock.Lock()
	defer sc.sqLock.Unlock()
	sc.errQueue = append(sc.errQueue, err)
	if sc.hasThread {
		if wg != nil {
			wg.Done()
		}
		return
	}
	sc.hasThread = true
	go sc.sendThread(wg)
}

func (sc *SendNb) sendThread(wg *sync.WaitGroup) {
	defer RecoverThread("send on error channel", func(err error) {
		if runt.IsSendOnClosedChannel(err) {
			return // ignore if the channel was or became closed
		}
		sc.panicFunc(err) // add to the lot
	})

	for {
		err := sc.getErr()
		if err == nil {
			break
		}
		if wg != nil {
			wg.Done() // signal read-to-send
			wg = nil
		}
		sc.SendChannel.Send(err) // may block and panic
	}
}

func (sc *SendNb) getErr() (err error) {
	sc.sqLock.Lock()
	defer sc.sqLock.Unlock()
	if len(sc.errQueue) == 0 {
		sc.hasThread = false
		return nil
	}
	err = sc.errQueue[0]
	copy(sc.errQueue[0:], sc.errQueue[1:])
	sc.errQueue = sc.errQueue[:len(sc.errQueue)-1]
	return
}
