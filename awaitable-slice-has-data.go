/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// hasData is bitfield determining if
// data exist begind inQ lock or outQ lock
type hasData struct {
	bits Atomic32[hasDataBitField]
}

// hasData returns true if there is data in the AwaitableSlice
//   - invoked at any time, no lock required to be held
func (h *hasData) hasData() (dataYes bool) {
	return h.bits.Load()&hasDataBit != 0
}

// setOutputLockEmpty clears hasDataBit if queueLockHasDataBit is
// still cleared
//   - invoked while holding queueLock
//   - invoked while hasDataBit is set
func (h *hasData) setOutputLockEmpty() (isEmpty bool) {
	for {

		// read initial state to be able to do CompareAndSwap
		var oldBits = h.bits.Load()
		if oldBits&inQHasDataBit != 0 {
			// queueLock received data, the queue will not become empty
			return // no hasDataBit clear return
		} else if h.bits.CompareAndSwap(oldBits, emptyNoBits) {
			return // set to empty return
		}

	}
}

// inQIsEmpty returns true if the AwaitableSlice is now empty
//   - isEmpty true: the entire queue is empty,
//     hasDataBit was reset using CompareAndSwap
//   - isEmpty false: there is data in inQ to be retrieved.
//     hasDataBits was not changed
//   - —
//   - invoked when outQ confirmed empty
//   - hasDataBit is known to be set
//   - purpose is to determine whether inQ lock must be accessed
//   - if inQ is also empty, the entire queue is empty
//   - invoked while holding outQ lock
func (h *hasData) inQIsEmpty() (isEmpty bool) {
	for {

		// read initial state to be able to do CompareAndSwap
		var oldBits = h.bits.Load()
		// if queueLock has no data, the entire queue is empty: isEmpty true
		isEmpty = oldBits&inQHasDataBit == 0
		// if the queue is not empty, return to access data behind qqueueLock
		if !isEmpty {
			return // data in queueLock return
		}
		// if queueLockHasDataBit is still zero, set hasDataBit to zero, too
		var newBits hasDataBitField = emptyNoBits // entire queue is empty
		if h.bits.CompareAndSwap(oldBits, newBits) {
			return // isEmpty true return
		}
	}
}

// resets while holding both locks
func (h *hasData) resetToHasDataBit() {
	h.bits.Store(hasDataBit)
}

func (h *hasData) resetToNoBits() {
	h.bits.Store(emptyNoBits)
}

func (h *hasData) resetToAllBits() {
	h.bits.Store(hasDataBit | inQHasDataBit)
}

const (
	// hasData indicates that the AwaitableSlice is not empty
	//	- bit 0 value 1
	hasDataBit hasDataBitField = 1 << iota
	// inQHasDataBit indicates that there is data behind queueLock
	//	- bit 1 value 2
	inQHasDataBit
	// emptyNoBits resets to initial zero-value state all bits cleared
	//	- trhe zero-value
	emptyNoBits hasDataBitField = 0
)

// [hasDataBit] [queueLockHasDataBit] [emptyNoBits]
type hasDataBitField uint32
