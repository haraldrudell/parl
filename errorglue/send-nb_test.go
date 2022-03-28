/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"sync"
	"testing"
	"time"
)

const snbShortTime = 10 * time.Microsecond

func TestSendNB(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	errCh := make(chan error)
	sendNb := SendNb{SendChannel: *NewSendChannel(errCh, nil)}
	var wg sync.WaitGroup
	wg.Add(1)
	sendNb.Send(err1, &wg)
	sendNb.Send(err2)

	wg.Wait() // wait until thread is about to send
	sendNb.sqLock.Lock()
	hasThread := sendNb.hasThread
	queueLen := len(sendNb.errQueue)
	sendNb.sqLock.Unlock()
	if !hasThread {
		t.Error("sendNB no thread")
	}
	if queueLen != 1 {
		t.Errorf("sendNB queue length not 1: %d", queueLen)
	}
	err := <-errCh
	if err != err1 {
		t.Errorf("Got wrong error: %v expected: %v", err, err1)
	}

	// cause a send panic
	close(errCh)
	// sendNb thread silently terminates
	// if things are bad, thread will panic…
	time.Sleep(snbShortTime)
	// termination is not indicated anywhere
}
