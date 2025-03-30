/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pslices/pslib"

// copyToSlice does slice-away from src, copying to dst and adding to *np
//   - src: pointer to source-slice
//   - dst: pointer to Read p-slice buffer
//   - np: pointer to Read n integer
//   - zeroOut missing: zero out is done
//   - zeroOut [NoZeroOut]: no zero-out
//   - isDone: true if dst was filled
//   - usable with [io.Read]-like functions reading from multiple slices
func CopyToSlice[T any](src, dst *[]T, np *int, zeroOut ...pslib.ZeroOut) (isDone bool) {

	// copy if anything to copy
	var d = *dst
	var sc = *src
	var nCopy = copy(d, sc)
	if nCopy == 0 {
		return // nothing to copy: isDone false
	}
	// items were copied

	// zero-out
	if len(zeroOut) == 0 || zeroOut[0] != NoZeroOut {
		clear(sc[:nCopy])
	}

	// update n *src *dst isDone
	*np += nCopy
	*src = sc[nCopy:]
	*dst = d[nCopy:]
	isDone = len(d) == nCopy

	return // bytes were copied return
}
