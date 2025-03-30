/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/pslices"
)

// list is a thread-safe structure fetauring:
//   - lock: making sliceList thread-safe
//   - sliceList: a slice-away list of value-slices
type list[T any] struct {
	// lock makes fields thread-safe and creates critical section
	// for accessing threads
	lock Mutex
	// sliceList is a queue of slices
	//	- no value-slice element of sliceList is empty or nil
	//	- sliceList is the result of a slice-expression on sliceList0.
	//		sliceList and sliceList0 share underlying array
	//	- sliceList is slice-away slice
	sliceList [][]T
	// sliceList0 is the full-length initial slice of the underlying array
	//	- the purpose of sliceList0 is to return sliceList to
	//		the beginning of the underlying array
	sliceList0 [][]T
	// isAlloc is true if sliceList0 is allocated
	isAlloc atomic.Bool
}

// isEmptyList returns true if sliceList is empty
func (q *list[T]) isEmptyList() (isEmpty bool) {
	return len(q.sliceList) == 0
}

// getListMetrics returns length and capacity of sliceList
func (q *list[T]) getListMetrics() (length, capacity int) {
	length = len(q.sliceList)
	capacity = cap(q.sliceList)
	return
}

// getListElementCount returns total number of elements in
// any value-slices in sliceList
func (q *list[T]) getListElementCount() (count int) {
	for i := range q.sliceList {
		count += len(q.sliceList[i])
	}
	return
}

// swapList gets sliceList and its ownership
//   - set may be nil
func (q *list[T]) swapList(set ...[][]T) (slices [][]T) {

	// set case
	if len(set) > 0 {
		var setSlices = set[0]
		q.isAlloc.Store(setSlices != nil)
		q.sliceList = setSlices
		if c := cap(setSlices); len(setSlices) < c {
			setSlices = setSlices[:c]
		}
		q.sliceList0 = setSlices
		return
	}

	// get q.sliceList and ownership
	slices = q.sliceList
	q.sliceList = nil
	q.sliceList0 = nil
	q.isAlloc.Store(false)

	return
}

// getListSlice gets sliceList but not its ownership
func (q *list[T]) getListSlice() (slices [][]T) {
	slices = q.sliceList
	return
}

// clearList empties sliceList
func (q *list[T]) clearList() {
	clear(q.sliceList)
	q.sliceList = q.sliceList0[:0]
}

// enqueueInList adds slice of value-slices to sliceList
func (q *list[T]) enqueueInList(slices ...[]T) {
	if len(slices) == 1 {
		pslices.SliceAwayAppend1(&q.sliceList, &q.sliceList0, slices[0], DoZeroOut)
		return
	}
	pslices.SliceAwayAppend(&q.sliceList, &q.sliceList0, slices, DoZeroOut)
}

// sliceAway retrieves a slice sliceList using slice-away
//   - if sliceList empty: nil
func (q *list[T]) dequeueFromList() (slice []T) {

	// empty case
	if len(q.sliceList) == 0 {
		return
	}

	// slice away
	slice = q.sliceList[0]
	q.sliceList[0] = nil
	if len(q.sliceList) == 1 {
		q.sliceList = q.sliceList0[:0]
	} else {
		q.sliceList = q.sliceList[1:]
	}

	return
}

// dequeueList moves all elements to dest
//   - dest must be of sufficient size
func (q *list[T]) dequeueList(dest []T) (n int) {

	// copy slices,deleting each copied slice
	for i := range len(q.sliceList) {
		// n0 is number of elements for this slice
		var n0 = copy(dest, q.sliceList[i])
		// because slice is set to nil, no clear is required
		q.sliceList[i] = nil
		n += n0
		dest = dest[n0:]
	}
	pslices.SetLength(&q.sliceList, 0)
	return
}

// dequeueNFromList moves elements for [io.Read] type functions
//   - dest: pointer to non-empty buffer.
//     Sliced away as elements are added
//   - np: pointer where the number of moved elements are added
//   - isDone: true if dest was filled completely
func (q *list[T]) dequeueNFromList(dest *[]T, np *int) (isDone bool) {
	for i := range len(q.sliceList) {
		var slicep = &q.sliceList[i]
		// copy with zero-out
		isDone = CopyToSlice(slicep, dest, np)
		// if the value-slice was copied empty, remove it
		if len(*slicep) == 0 {
			*slicep = nil
			if len(q.sliceList) == 1 {
				q.sliceList = q.sliceList0[:0]
			} else {
				q.sliceList = q.sliceList[1:]
			}
		}
		if isDone {
			return
		}
	}
	return
}

// getLastSliceMetrics provides information on last slice of sliceList
func (q *list[T]) getLastSliceMetrics() (length, capacity int) {
	if len(q.sliceList) == 0 {
		return
	}
	var slice = q.sliceList[len(q.sliceList)-1]
	length = len(slice)
	capacity = cap(slice)
	return
}

func (q *list[T]) getLastSliceSlice() (lastSlice []T) {
	if len(q.sliceList) == 0 {
		return
	}
	lastSlice = q.sliceList[len(q.sliceList)-1]
	return
}

// swapLastSlice gets or sets last slice
//   - on get, lastSlice remains
func (q *list[T]) swapLastSlice(slice ...[]T) (lastSlice []T) {

	// pointer to last slice
	var slicep *[]T
	if len(q.sliceList) > 0 {
		slicep = &q.sliceList[len(q.sliceList)-1]
	}

	// get case
	if len(slice) == 0 {
		if slicep != nil {
			lastSlice = *slicep
		}
		return
	}

	// put case
	if slicep != nil {
		*slicep = slice[0]
		return
	}
	q.sliceList = append(q.sliceList, slice[0])

	return
}

// enqueueInLastSlice appends value to the last slice in sliceList
//   - value: non-empty value list
//   - slice list must have been verified to not be empty
func (q *list[T]) enqueueInLastSlice(value ...T) {

	// index of last slice in sliceList
	var index = len(q.sliceList) - 1
	var slicep = &q.sliceList[index]
	// the last slice from sliceList
	var slice = *slicep
	var length = len(slice)

	// if value fits, copy it into slice
	if n := length + len(value); n <= cap(slice) {
		slice = slice[:n]
		if len(value) == 1 {
			slice[length] = value[0]
		} else {
			copy(slice[length:], value)
		}
		*slicep = slice
		return
	}

	// realloc
	*slicep = append(slice, value...)
}
