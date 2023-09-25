/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

type SymlinkList struct {
	symlinkTargets []string
}

func NewSymlinkList() (list *SymlinkList) {
	return &SymlinkList{}
}
