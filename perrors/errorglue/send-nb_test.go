/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"testing"
)

func TestSendNB(t *testing.T) {
	var err1 = errors.New("err1")
	var err2 = errors.New("err2")
	var errCh = make(chan error)

	var err error
	var hasThread bool
	var queueLen int

	sendNb := NewSendNb(errCh)

	sendNb.Send(err1)
	sendNb.Send(err2)
	err = <-errCh
	if err != err1 {
		t.Errorf("Got wrong error: %v expected: %v", err, err1)
	}
	sendNb.sqLock.Lock()
	hasThread = sendNb.hasThread
	sendNb.sqLock.Unlock()
	if !hasThread {
		t.Error("sendNB no thread")
	}

	if sendNb.IsShutdown() {
		t.Error("sendNb.IsShutdown")
	}
	sendNb.Shutdown()
	if !sendNb.IsShutdown() {
		t.Error("sendNb.IsShutdown false")
	}
	sendNb.sqLock.Lock()
	hasThread = sendNb.hasThread
	queueLen = len(sendNb.sendQueue)
	sendNb.sqLock.Unlock()
	if hasThread {
		t.Error("sendNB has thread")
	}
	if queueLen != 0 {
		t.Errorf("sendNB queue length not 0: %d", queueLen)
	}
}
