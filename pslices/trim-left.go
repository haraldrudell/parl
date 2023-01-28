/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// TrimLeft removes count bytes from the start of slicep, copying remaining bytes to its beginning.
package pslices

// TrimLeft removes count bytes from the start of slicep, copying remaining bytes to its beginning.
//   - slicep’s length is adjusted
//   - if count < 1 or slicep is empty or nil, nothing is done
//   - if count >= len(slicep) slicep is emptied
//   - no allocation or free is triggered
func TrimLeft[E any](slicep *[]E, count int, noZero ...bool) {

	// get valid length and count
	s := *slicep
	length := len(s)
	if count < 1 || length == 0 {
		return // nothing to do return
	} else if count > length {
		count = length
	}

	// delete the count first element from slice of length length
	//	- count is 1…length
	//	- length is len(*slicep)
	if count < length {
		// move the element at the end that will be kept
		copy(s, s[count:])
	}

	// zero out deleted element at end
	newLength := length - count
	doZero := true
	if len(noZero) > 0 {
		doZero = !noZero[0]
	}
	if doZero {
		var e E
		for i := newLength; i < length; i++ {
			s[i] = e
		}
	}

	// adjust slice length
	*slicep = s[:newLength]
}
