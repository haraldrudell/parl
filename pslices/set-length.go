/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

// SetLength adjusts the lenght of *slicep extending with append if necessary.
//   - slicep’s length is adjusted
//   - if newLength > cap, slice may be reallocated
func SetLength[E any](slicep *[]E, newLength int, noZero ...bool) {

	s := *slicep
	length := len(s)
	if newLength == length {
		return // same length nothing to do return
	}

	doZero := true
	if len(noZero) > 0 {
		doZero = !noZero[0]
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
			if doZero {
				var e E
				for i := length; i < cap; i++ {
					s[i] = e
				}
			}
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
