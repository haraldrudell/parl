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
	defer a.outputLock.Unlock()
	defer a.queueLock.Unlock()
	a.outputLock.Lock()
	a.queueLock.Lock()

	var sL []string

	// behind queueLock
	sL = append(sL, Sprintf("hasData: %t q %s slices %s loc %t cached %d",
		// “hasData: false”
		a.hasData.Load(),
		// queue: “0(10)”
		printSlice(a.queue),
		// slices, slices0, isLocalSlice, cachedInput
		// “slices 0(cap0/0 tot0 offs-1) loc false cached 10”
		printSlice2Away(a.slices, a.slices0), a.isLocalSlice, cap(a.cachedInput),
	))

	// behind outputLock
	sL = append(sL, Sprintf("  out %s outs %s cached %d",
		// output output0: “out 0(cap 9/10 offs 1)”
		printSliceAway(a.output, a.output0),
		// outputs outputs0 cachedOutput
		// “outs 0(cap0/0 tot0 offs-1) cached 10”
		printSlice2Away(a.outputs, a.outputs0), cap(a.cachedOutput),
	))

	// data Wait
	sL = append(sL, Sprintf("  data %s end %s",
		// dataWait: “act-cls-chc:false-false-false”
		printCyclic(&a.dataWait),
		// end: “act-cls-chc:false-false-false”
		printCyclic(&a.emptyWait),
	))

	s = strings.Join(sL, "\n")
	return
}

func printCyclic(c *LazyCyclic) (s2 string) {
	var isClosed bool
	select {
	case <-c.Cyclic.Ch():
		isClosed = true
	default:
	}
	return fmt.Sprintf("act-cls-chc:%t-%t-%t",
		// whether active
		c.IsActive.Load(),
		// closed flag
		c.Cyclic.IsClosed(),
		// channel actually closed
		isClosed,
	)
}

func printSlice[T any](s []T) (s2 string) {
	return fmt.Sprintf("%d(%d)", len(s), cap(s))
}

func printSliceAway[T any](s, s0 []T) (s2 string) {
	var offset, hasValue = pslices.Offset(s0, s)
	if !hasValue {
		offset = -1
	}
	return fmt.Sprintf("%d(cap%d/%d offs%d)",
		len(s), cap(s), cap(s0), offset,
	)
}

func printSlice2Away[T any](s, s0 [][]T) (s2 string) {
	var offset, hasValue = pslices.Offset(s0, s)
	if !hasValue {
		offset = -1
	}
	var total int
	for _, sx := range s {
		total += len(sx)
	}
	return fmt.Sprintf("%d(cap%d/%d tot%d offs%d)",
		len(s), cap(s), cap(s0), total, offset,
	)
}
