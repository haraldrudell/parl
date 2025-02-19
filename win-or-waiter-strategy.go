/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/sets"

const (
	// WinOrWaiterAnyValue allows a thread to accept any calculated value
	WinOrWaiterAnyValue WinOrWaiterStrategy = iota
	// WinOrWaiterMustBeLater forces a calculation commencing after a thread arrives
	//	- WinOrWaiter caclulations are serialized, ie. a new calculation does not start prior to
	//		the conclusion of the previous calulation
	//	- thread arrival time is prior to acquiring the lock
	WinOrWaiterMustBeLater
)

type WinOrWaiterStrategy uint8

func (ws WinOrWaiterStrategy) String() (s string) {
	return winOrWaiterSet.StringT(ws)
}

func (ws WinOrWaiterStrategy) IsValid() (isValid bool) {
	return winOrWaiterSet.IsValid(ws)
}

var winOrWaiterSet = sets.NewSet[WinOrWaiterStrategy]([]sets.SetElement[WinOrWaiterStrategy]{
	{ValueV: WinOrWaiterAnyValue, Name: "anyValue"},
	{ValueV: WinOrWaiterMustBeLater, Name: "mustBeLater"},
})
