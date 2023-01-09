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
func TrimLeft[E any](slicep *[]E, count int) {
	length := len(*slicep)
	if count < 1 || length == 0 {
		return // nothing to do return
	} else if count > length {
		count = length
	}
	if count < length {
		copy(*slicep, (*slicep)[count:])
	}
	*slicep = (*slicep)[:length-count]
}
