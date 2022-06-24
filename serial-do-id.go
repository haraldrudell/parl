/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type SerialDoID string

var serialDoID UniqueID[SerialDoID]

func (id SerialDoID) String() (s string) {
	return string(id)
}
