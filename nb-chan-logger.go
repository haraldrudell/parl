/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// [NBChanLogger] logging thread does not close until channel is closed
	NBChanExpectClose = true
	// [NBChanLogger] logging thread exits when channel has been read to empty
	NBChanWillNotClose = false
)

// generates labels "1"… separating different channel instances
var nbChanLoggerID UniqueIDUint64

// NBChanLogger is a debug logger for an NBChan instance
//   - label is a string leading printouts, default a small integer
//   - NBChan is the channel watched
//   - printout continues until the channel is empty and the thread has exited
//   - if expectClose is true, printout will continue until the underlying channel is closed
//   - log default is parl.Log
func NBChanLogger[T any](label string, n *NBChan[T], expectClose bool, log ...PrintfFunc) {
	if n == nil {
		panic(perrors.NewPF("nbChan cannot be nil"))
	}
	if label == "" {
		label = strconv.Itoa(int(nbChanLoggerID.ID()))
	}
	var log0 PrintfFunc
	if len(log) > 0 {
		log0 = log[0]
	}
	if log0 == nil {
		log0 = Log
	}
	go doLogging(label, n, expectClose, log0)
}

// doLogging prints NBChan status output every second
func doLogging[T any](label string, n *NBChan[T], expectClose bool, log PrintfFunc) {

	// ticker for periodic printing
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()

	var endCh = n.WaitForCloseCh()
	for {
		log(label + "\x20" + NBChanState(n))

		if n.Count() == 0 &&
			n.ThreadStatus() == NBChanExit &&
			n.sends.Load() == 0 &&
			n.gets.Load() == 0 &&
			(!expectClose || n.DidClose()) {
			return
		}

		select {
		case <-endCh:
			endCh = nil
		case <-ticker.C:
		}
	}
}

// “length/i/o: 1/0/0 close-now:true-false thread: send sends: 0 gets: 0 always: true-true chClosed: false err: false”
func NBChanState[T any](n *NBChan[T]) (s string) {
	n.inputLock.Lock()
	var in = len(n.inputQueue)
	n.inputLock.Unlock()
	n.outputLock.Lock()
	var out = len(n.outputQueue)
	n.outputLock.Unlock()
	var threadType string
	if n.noThread.Load() {
		threadType = "-" + NBChanNone.String()
	} else if n.isThreadAlways.Load() {
		threadType = "-" + NBChanAlways.String()
	}
	var alertValue string
	if len(n.threadCh) > 0 {
		alertValue = "-alertValue"
	}
	var hasData string
	if n.isDataAvailable.Load() {
		hasData = "data"
	} else {
		hasData = "empty"
	}

	return Sprintf("length/i/o:%d/%d/%d-%s%s sends/gets:%d/%d thread:%s%s close/now/ch:%t/%t/%t err: %t",

		// “length/i/o: 1/0/0” unsentCount, len input, len output
		n.Count(), in, out, hasData, alertValue,

		// “send sends: 0 gets: 0” pending Send/SendMany, Get
		n.sends.Load(), n.gets.Load(),

		// “thread: chSend” ThreadStatus
		n.ThreadStatus(), threadType,

		// “close-now:true-false” Close-CloseNow
		n.isCloseInvoked.Load(), n.isCloseNow.IsInvoked(), n.IsClosed(),

		// “err: false” if NBCHan had panic or close error
		n.GetError() != nil,
	)
}
