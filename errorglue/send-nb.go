/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"os"
	"sync"

	"github.com/haraldrudell/parl/pruntime"
)

// SendNb implements non-blocking send using a thread and buffer up to size of int
type SendNb struct {
	SendChannel
	sqLock    sync.Mutex
	sendQueue []error // inside lock
	hasThread bool    // inside lock
}

// Send sends an error on the error channel. Non-blocking. Thread-safe.
// wg can be used to wait until the thread is near-send
func (sc *SendNb) Send(err error, wgs ...*sync.WaitGroup) {
	if err == nil {
		return // nothing to send
	}

	// get possible WaitGroup
	var wg *sync.WaitGroup
	if len(wgs) > 0 {
		wg = wgs[0]
	}
	if sc.errCh == nil {
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
	sc.sendQueue = append(sc.sendQueue, err) // put err in send queue
	if sc.hasThread {                        // the thread is already blocked
		if wg != nil {
			wg.Done()
		}
		return // err put in queue
	}

	// launch thread
	sc.hasThread = true
	go sc.sendThread(wg) // send err in new thread
}

func (sc *SendNb) sendThread(wg *sync.WaitGroup) {
	defer RecoverThread("send on error channel", func(err error) {
		if pruntime.IsSendOnClosedChannel(err) {
			return // ignore if the channel was or became closed
		}

		// no other panics should occur — we do not know what this is
		// This will never happen. This is the best we can do when it does
		fmt.Fprintln(os.Stderr, err)
	})

	for {
		err := sc.getErrFromSendQueue()
		if err == nil {
			break // end of queue: shutdown thread
		}
		if wg != nil {
			wg.Done() // signal near-send
			wg = nil
		}
		sc.SendChannel.Send(err) // may block and panic
	}
}

func (sc *SendNb) getErrFromSendQueue() (err error) {
	sc.sqLock.Lock()
	defer sc.sqLock.Unlock()
	if len(sc.sendQueue) == 0 {
		sc.hasThread = false
		return nil
	}
	err = sc.sendQueue[0]
	copy(sc.sendQueue[0:], sc.sendQueue[1:])
	sc.sendQueue = sc.sendQueue[:len(sc.sendQueue)-1]
	return
}
