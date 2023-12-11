/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import "io/fs"

// compareDirEntry sorts [fs.DirEntry] by basename
//   - ascending 8-bit character
func compareDirEntry(a, b fs.DirEntry) (result int) {
	var an = a.Name()
	var bn = b.Name()
	if an < bn {
		return -1
	} else if an > bn {
		return 1
	}
	return
}
