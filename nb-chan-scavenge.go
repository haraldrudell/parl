/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync/atomic"

// Scavenge attempts to set capacity of the two queues to setCapacity
//   - Scavenge allows for reducing queue capacity thus reduce memory leaks
//   - queue capacities are reduced to the setCapacity value,
//     while ensuring enough capacity for current elements
//   - unused queue elements are set to zero-value to prevent memory leaks
func (n *NBChan[T]) Scavenge(setCapacity int) {
	// holding [NBChan.outputLock] prevents queue swap
	n.outputLock.Lock()
	defer n.outputLock.Unlock()

	var capacity = n.reduceQueue(&n.outputQueue, setCapacity)
	atomic.StoreUint64(&n.outputCapacity, uint64(capacity))

	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	capacity = n.reduceQueue(&n.inputQueue, setCapacity)
	atomic.StoreUint64(&n.inputCapacity, uint64(capacity))
}

// reduceQueue reduces the capacity of a queue to avoid memory leaks
func (n *NBChan[T]) reduceQueue(queuep *[]T, setCapacity int) (capacity int) {

	// check for small or unallocated queue
	var q = *queuep
	capacity = cap(q)
	if capacity <= setCapacity {
		return // unallocated or small enough queue return
	}

	// ensure enough for length
	var length = len(q)
	if length > setCapacity {
		if setCapacity = length; setCapacity == capacity {
			return // queue capacity cannot be reduced return
		}
	}

	// reduce queue capacity
	var newQueue = make([]T, setCapacity)
	copy(newQueue, q)
	*queuep = newQueue[:length]
	capacity = cap(newQueue)

	return
}
