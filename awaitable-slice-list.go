/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/pslices"
	"github.com/haraldrudell/parl/pslices/pslib"
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

// setList sets sliceList, typically a one-time operation
//   - sliceList non-nil: provides allocated capacity to the list instance
//   - — if sliceList is not empty, it cannot contain nil or empty slice values
//   - — sliceList must have non-zero capacity
//   - sliceList nil: deallocates sliceList
func (q *list[T]) setList(sliceList [][]T) {
	q.sliceList = sliceList
	if c := cap(sliceList); len(sliceList) < c {
		sliceList = sliceList[:c]
	}
	q.sliceList0 = sliceList
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

// enqueueInList adds value-slices to sliceList
func (q *list[T]) enqueueInList(slices ...[]T) {
	if len(slices) == 1 {
		pslices.SliceAwayAppend1(&q.sliceList, &q.sliceList0, slices[0], DoZeroOut)
		return
	}
	pslices.SliceAwayAppend(&q.sliceList, &q.sliceList0, slices, DoZeroOut)
}

// sliceAway returns a value-slice from sliceList using slice-away
//   - slice non-nil: a non-empty value-slice removed from sliceList
//   - slice nil: sliceList is empty
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
//   - dest: must be of sufficient size
//   - — if dest is too short, elements are lost
//   - n: number of elements copied to dest
func (q *list[T]) dequeueList(dest []T) (n int) {

	// copy the elements of all slices in sliceList to dest
	for i := range len(q.sliceList) {

		// n0 is number of elements copied from this value-slice
		var n0 = copy(dest, q.sliceList[i])

		n += n0
		dest = dest[n0:]
	}

	// set sliceList length to zero with zero-out
	pslices.SetLength(&q.sliceList, 0, pslib.DoZeroOut)

	return
}

// dequeueNFromList moves elements for [io.Read] type functions
//   - dest: pointer to non-empty buffer.
//     Sliced away as elements are added
//   - np: pointer where the number of moved elements are added
//   - isDone: true if dest was filled completely
func (q *list[T]) dequeueNFromList(dest *[]T, np *int, zeroOut pslib.ZeroOut) (isDone bool) {

	// process each slice in sliceList in order
	//	- sliceList is [][]T
	//	- each value-slice from sliceList:
	//	- — type []T
	//	- — non-nil with non-zero length and capacity
	for len(q.sliceList) > 0 {
		// slicep *[]T is non-zero length
		var slicep = &q.sliceList[0]
		// sliceBefore is source slice before slice-away operation
		var sliceBefore = *slicep

		// copy slicep to dest without zero-out
		isDone = CopyToSlice(slicep, dest, np, pslib.NoZeroOut)

		// if the value-slice was copied empty, remove it
		if slicepLength := len(*slicep); slicepLength == 0 {
			// set to nil, zero-out not required
			*slicep = nil
			// slice-away
			if len(q.sliceList) == 1 {
				q.sliceList = q.sliceList0[:0]
			} else {
				q.sliceList = q.sliceList[1:]
			}
		} else if zeroOut == pslib.DoZeroOut {
			if copyCount := len(sliceBefore) - slicepLength; copyCount > 0 {
				clear(sliceBefore[:copyCount])
			}
		}

		// isDone true means dest is now empty
		if isDone {
			return
		}
	}

	return
}

// getLastSliceMetrics provides information on last slice of sliceList
//   - length: non-zero length of last slice if it exists
//   - capacity: non-zero capacity of last slice if it exists
func (q *list[T]) getLastSliceMetrics() (length, capacity int) {

	// empty sliceList case
	if len(q.sliceList) == 0 {
		return
	}

	var slice = q.sliceList[len(q.sliceList)-1]
	length = len(slice)
	capacity = cap(slice)
	return
}

// getLastSliceSlice returns the last slice without ownership if it exists
//   - lastSlice non-nil: a non-empty value-slice
//   - lastSlice nil: sliceList is empty
func (q *list[T]) getLastSliceSlice() (lastSlice []T) {

	// sliceList empty case
	if len(q.sliceList) == 0 {
		return
	}

	lastSlice = q.sliceList[len(q.sliceList)-1]
	return
}

// swapLastSlice gets or sets last slice
//   - slice empty: return last slice but not its ownership
//   - slice[0] present: must be slice of non-zero length
//   - — if last slice exists, it is overwritten by slice[0]
//   - — if sliceList is empty, slice[0] is appended as new last slice
func (q *list[T]) swapLastSlice(slice ...[]T) (lastSlice []T) {

	// slicep is pointer to last slice or nil if it does not exist
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

// enqueueInLastSlice appends values to the last slice in sliceList
//   - values: non-empty value list
//   - sliceList must have been verified to not be empty
//   - operation may be copy or re-allocating append
func (q *list[T]) enqueueInLastSlice(values ...T) {

	// index of last slice in sliceList
	var index = len(q.sliceList) - 1
	// pointer to last slice
	var slicep = &q.sliceList[index]
	// the last slice in sliceList
	var slice = *slicep
	// initial length of last slice
	var length = len(slice)

	// if value fits, copy it into slice
	if n := length + len(values); n <= cap(slice) {
		// extend slice to new length
		slice = slice[:n]
		if len(values) == 1 {
			slice[length] = values[0]
		} else {
			// copy values to last part of slice
			copy(slice[length:], values)
		}
		// update last slice
		*slicep = slice
		return
	}

	// realloc
	*slicep = append(slice, values...)
}
