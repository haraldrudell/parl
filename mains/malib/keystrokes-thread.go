/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pdebug"
	"github.com/haraldrudell/parl/pruntime"
)

// D is hook to facilitate debug printing
var D parl.PrintfFunc

// KeystrokesThread is a minimal goroutine unable to exit due to read from [os.Stdin]
//   - isActive: true while background is still monitoring standard input
//   - — when set to false, thread will exit as soon as practical
//   - errCh: errors and panics are sent on errCh, should be size 2
//   - — errCh closes on thread exit making the thread awaitable
//   - slicep: where slices of bytes are appended, behind sliceLock
//   - sliceLock: makes slicep thread-safe
//   - dataCh: if empty, thread sends a value when data becomes available
//   - —
//   - events causing thread-exit: errCh closing
//   - — stdin closing: ^D: StdinErr with io.EOF is sent on errCh
//   - — panic: StdinErr with RecoverAny non-nil is sent on errCh
//   - — error from os.Stdin: StdinErr is sent on errCh
//   - — isActive being false
//   - only takes Go types to reduce memory references
func KeystrokesThread(isActive *atomic.Bool, errCh chan error, slicep *[][]byte, sliceLock *sync.Mutex, dataCh chan struct{}) {
	defer close(errCh)
	var err error
	defer keyRecovery(isActive, errCh, &err)

	var n int
	var p = make([]byte, stdinBufSize)
	if D != nil {
		D("%s at for:\n%s\n—",
			codeLine.FuncName(),
			pdebug.NewStack(0),
		)
	}
	for {

		// read from os.Stdin
		//	- blocks here
		n, err = os.Stdin.Read(p)

		// check if still active
		if !isActive.Load() {
			return // background no longer receiving data
		}

		// send any data
		if n > 0 {
			// clone buffer bytes
			var sliceClone = slices.Clone(p[:n])
			appendToSlice(sliceClone, slicep, sliceLock)
			// trigger data-wait on dataCh
			if len(dataCh) == 0 {
				dataCh <- struct{}{}
			}
		}

		// handle any errors
		if err != nil {
			if D != nil {
				D("%s err: %s",
					codeLine.FuncName(), err,
				)
			}
			errCh <- &StdinErr{Err: err}
			return
		}
	}
}

// keyRecovery recovers any panic
//   - isActive: true if background is still monitoring the thread
//   - errCh: error channel to background
func keyRecovery(isActive *atomic.Bool, errCh chan error, errp *error) {

	// recover any panic
	var reccoverAny = recover()
	var isPanic = reccoverAny != nil
	var panicS string
	if D != nil {
		D("%s exiting: isPanic: %t err: %v isActive: %t",
			codeLine.FuncName(),
			isPanic, *errp, isActive.Load(),
		)
	}
	if !isPanic && isActive.Load() {
		return // no panic, background still monitoring
	}

	// send any panic
	if isPanic {
		panicS = fmt.Sprintf("panic non-error value %T “%[1]v”", reccoverAny)
		var s = &StdinErr{RecoverAny: reccoverAny}
		if e, isError := reccoverAny.(error); isError {
			s.Err = e
		} else {
			s.Err = errors.New(panicS)
		}
		errCh <- s
	}

	// if monitoring still present: done
	if isActive.Load() {
		return // panic sent and monitored
	}

	// print to standard error
	var eS string
	if isPanic {
		eS = panicS
	} else if err := *errp; err != nil {
		eS = fmt.Sprintf("stdin: error: “%s”", err)
	} else {
		eS = "stdin exits after monitoring stopped"
	}
	fmt.Fprintln(os.Stderr, eS)
}

// appendToSlice appends using lock
//   - slice: a slice to append
//   - *slicep: pointer to slice list
//   - sliceLock: lock for slicep
func appendToSlice(slice []byte, slicep *[][]byte, sliceLock *sync.Mutex) {
	sliceLock.Lock()
	defer sliceLock.Unlock()

	*slicep = append(*slicep, slice)
}

// KeyStrokes stack trace 250404:
// ID: 17 status: ‘running’
// github.com/haraldrudell/parl/mains/malib.KeystrokesThread(0x140000a8074, 0x140000a2000, 0x140000a8088, 0x140000a8080, 0x140000a6000)
//   /opt/sw/parl/mains/malib/keystrokes-thread.go:41
// Parent-ID: 7 go: github.com/haraldrudell/parl/mains/malib.(*StdinReader).createThread
//   /opt/sw/parl/mains/malib/stdin-reader.go:171

const (
	// size of Stdin Read buffer
	stdinBufSize = 1024
)

// codeLine caches KeystrokesThread function identifier
var codeLine pruntime.CachedLocation
