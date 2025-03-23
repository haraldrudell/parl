/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "github.com/haraldrudell/parl/pslices/pslib"

// SetLength adjusts the lenght of *slicep extending with append if necessary
//   - slicep: pointer to slice whose length is adjusted
//   - newLength : the length of *slicep on return
//   - — if newLength > cap, slice may be reallocated
//   - noZero missing or DoZeroOut: elements becoming unused are set to zero-value
//   - — if element contain pointers, such elements are a temporary memory leak
//   - noZero NoZeroOut: no zero-out of elements
func SetLength[E any](slicep *[]E, newLength int, noZero ...pslib.ZeroOut) {

	s := *slicep
	length := len(s)
	if newLength == length {
		return // same length nothing to do return
	}

	doZero := true
	if len(noZero) > 0 {
		doZero = noZero[0] != pslib.NoZeroOut
	}

	// shortening slice case
	if newLength < length {
		// zero out removed elements
		if doZero {
			var e E
			for i := newLength; i < length; i++ {
				s[i] = e
			}
		}
		(*slicep) = s[:newLength]
		return // slice shortened return
	}

	// extending with append case
	if cap := cap(s); newLength > cap {
		if length < cap {
			s = s[:cap] // extend up to cap
		}
		*slicep = append(s, make([]E, newLength-cap)...)
		return
	}

	// extending length within cap
	s = s[:newLength]
	if doZero {
		var e E
		for i := length; i < newLength; i++ {
			s[i] = e
		}
	}
	*slicep = s
}
