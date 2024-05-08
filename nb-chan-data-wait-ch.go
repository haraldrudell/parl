/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// updateDataAvailable obtains a properly configured dataWaitCh
//   - used by DataWaitCh to obtain the channel
//   - by thread after sending an item
//   - by thread when exiting from CloseNow
func (n *NBChan[T]) updateDataAvailable() (dataCh chan struct{}) {
	if n.closableChan.IsClosed() {
		return n.setDataAvailableAfterClose()
	}
	return n.setDataAvailable(n.unsentCount.Load() > 0)
}

// setDataAvailableAfterClose ensures dataWaitCh is initialized and triggered
//   - consumers waiting for dataWaitCh will then not block
//   - used at end of Close, CloseNow and deferred close by thread
func (n *NBChan[T]) setDataAvailableAfterClose() (dataCh chan struct{}) {
	return n.setDataAvailable(true)
}

// setDataAvailable configured dataWaitCh to be aligned with isAvailable
//   - updated by Get Send SendMany
//   - also indirect by updateDataAvailable and setDataAvailableAfterClose
func (n *NBChan[T]) setDataAvailable(isAvailable bool) (dataCh chan struct{}) {
	if chp := n.dataWaitCh.Load(); chp != nil && n.isDataAvailable.Load() == isAvailable {
		dataCh = *chp
		return // initialized and in correct state return: noop
	}
	n.availableLock.Lock()
	defer n.availableLock.Unlock()

	// not yet initialized case
	var chp = n.dataWaitCh.Load()
	if chp == nil {
		dataCh = make(chan struct{})
		if isAvailable {
			close(dataCh)
		}
		n.dataWaitCh.Store(&dataCh)
		n.isDataAvailable.Store(isAvailable)
		return // channel initialized and state set return
	}

	// is state correct?
	dataCh = *chp
	if n.isDataAvailable.Load() == isAvailable {
		return // channel was initialized and state was correct return
	}

	// should channel be closed: if data is available
	if isAvailable {
		close(*chp)
		n.isDataAvailable.Store(true)
		return // channel closed andn state updated return
	}

	// replace with open channel: data is not available
	dataCh = make(chan struct{})
	n.dataWaitCh.Store(&dataCh)
	n.isDataAvailable.Store(false)
	return // new open channel stored and state updated return
}
