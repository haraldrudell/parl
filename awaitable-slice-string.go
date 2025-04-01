/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/pslices"
)

func AwaitableSliceString[T any](a *AwaitableSlice[T]) (s string) {
	defer a.outQ.lock.Lock().Unlock()
	defer a.outQ.InQ.lock.Lock().Unlock()

	var sL []string

	// behind queueLock
	sL = append(sL, Sprintf(
		"hasData: %t q %s slices %s loc %t cached %d",
		// “hasData: false”
		a.outQ.HasDataBits.bits.Load(),
		// queue: “q 0(10)”
		printSlice(a.outQ.InQ.primary),
		// slices, slices0, isLocalSlice, cachedInput
		// “slices 0(cap0/0 tot0 offs-1) loc false cached 10”
		printSlice2Away(a.outQ.sliceList, a.outQ.sliceList0), cap(a.outQ.InQ.cachedInput),
	))

	// behind outputLock
	sL = append(sL, Sprintf(
		"  out %s outs %s cached %d",
		// output output0: “out 0(cap 9/10 offs 1)”
		printSliceAway(a.outQ.head, a.outQ.head0),
		// outputs outputs0 cachedOutput
		// “outs 0(cap0/0 tot0 offs-1) cached 10”
		printSlice2Away(a.outQ.sliceList, a.outQ.sliceList0), cap(a.outQ.cachedOutput),
	))

	// data Wait
	var isEmptyWait = a.isCloseInvoked.Load()
	var isEmpty bool
	if isEmptyWait {
		isEmpty = a.isEmpty.IsClosed()
	}
	sL = append(sL, Sprintf(
		"  data %s end act-cls:%t-%t",
		// dataWait: “data act-cls-chc:false-false-false”
		printCyclic(&a.dataWait),
		// end: “end act-cls:false-false”
		isEmptyWait, isEmpty,
	))

	s = strings.Join(sL, "\n")
	return
}

// printCyclic returns cyclic state “act-cls-chc:false-false-false”
func printCyclic(c *LazyCyclic) (s2 string) {
	var isActive, isClosed, isChannelClosed bool
	if isActive = c.IsActive.Load(); isActive {
		isClosed = c.Cyclic.IsClosed()
		select {
		case <-c.Cyclic.Ch():
			isChannelClosed = true
		default:
		}
	}
	return fmt.Sprintf(
		"act-cls-chc:%t-%t-%t",
		// whether active
		isActive,
		// is-closed flag
		isClosed,
		// channel actually closed
		isChannelClosed,
	)
}

// printSlice returns slice state “1(10)”
func printSlice[T any](s []T) (s2 string) { return fmt.Sprintf("%d(%d)", len(s), cap(s)) }

// printSliceAway returns slice-away state “0(cap 9/10 offs 1)”
func printSliceAway[T any](s, s0 []T) (s2 string) {
	var offset, hasValue = pslices.Offset(s0, s)
	if !hasValue {
		offset = -1
	}
	return fmt.Sprintf(
		"%d(cap%d/%d offs%d)",
		len(s), cap(s), cap(s0), offset,
	)
}

// printSlice2Away returns slice-of-slices slice-away state “0(cap0/0 tot0 offs-1)”
func printSlice2Away[T any](s, s0 [][]T) (s2 string) {
	var offset, hasValue = pslices.Offset(s0, s)
	if !hasValue {
		offset = -1
	}
	var total int
	for _, sx := range s {
		total += len(sx)
	}
	return fmt.Sprintf(
		"%d(cap%d/%d tot%d offs%d)",
		len(s), cap(s), cap(s0), total, offset,
	)
}
