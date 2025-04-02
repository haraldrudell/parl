/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// hasData is bitfield determining if
// data exists behind inQ lock or outQ lock
//   - purpose is to provide hasData with integrity at all times and
//   - to determine if data exists behind inQ
//     requiring that lock to be acquired
type hasData struct {
	bits Atomic32[hasDataBitField]
}

// hasData returns true if there is data in the AwaitableSlice
//   - invoked at any time, no lock required to be held
func (h *hasData) hasData() (dataYes bool) {
	return h.bits.Load()&hasDataBit != 0
}

// setOutputLockEmpty clears hasDataBit if inQ-hasdata is
// still cleared
//   - purpose is to clear all bits when outQ is about to be emptied,
//     but only if no new data has enetered inQ
//   - invoked while holding outQ lock
//   - invoked while hasDataBit is set
func (h *hasData) setOutputLockEmpty() {
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

// isInQEmpty returns true if the AwaitableSlice is now empty
//   - isEmpty true: inQ lock does not hacve to be acquired,
//     there is no data behind it.
//     hasDataBit was reset using CompareAndSwap.
//     All bits were cleared
//   - isEmpty false: there is data in inQ lock to be retrieved.
//     hasDataBits was not changed
//   - —
//   - purpose is to determine whether inQ lock must be accessed
//   - invoked when outQ confirmed empty (Get GetSlice Read)
//     or known to become empty (GetSlices GetAll)
//   - hasDataBit is known to be set
//   - if inQ is also empty, the entire queue is empty
//   - must hold outQ lock
func (h *hasData) isInQEmpty() (isEmpty bool) {
	for {

		// read initial state to be able to do CompareAndSwap
		var oldBits = h.bits.Load()
		// if inQ lock has no data, the entire queue is empty: isEmpty true
		isEmpty = oldBits&inQHasDataBit == 0
		// not isEmpty means inQ lock has data
		//	- if inQ is not empty, outQ lock holder must acquire inQ lock to retrieve its data
		if !isEmpty {
			return // data in queueLock return: isEmpty false
		}
		// if queueLockHasDataBit is still zero, set hasDataBit to zero, too
		var newBits hasDataBitField = emptyNoBits // entire queue is empty
		if h.bits.CompareAndSwap(oldBits, newBits) {
			return // isEmpty true return
		}
	}
}

// resetToHasDataBit sets hasData and clears inQ data
//   - must hold both inQ lock and outQ lock
//   - must know that inQ will be emptied
func (h *hasData) resetToHasDataBit() {
	h.bits.Store(hasDataBit)
}

// setAllBits sets hasData and inQ-has-data
//   - must hold inQ lock
//   - must know that inQ is not empty
func (h *hasData) setAllBits() {
	if h.bits.Load() == hasDataBit|inQHasDataBit {
		return
	}
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
