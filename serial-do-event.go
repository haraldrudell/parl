/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"time"
)

type SerialDoEvent struct {
	ID SerialDoID
	SerialDoType
	time.Time
	*SerialDo
}

func NewSerialDoEvent(typ SerialDoType, t time.Time, serialDo *SerialDo) (event *SerialDoEvent) {
	return &SerialDoEvent{
		ID:           serialDoID.ID(),
		SerialDoType: typ,
		Time:         t,
		SerialDo:     serialDo,
	}
}

func (e *SerialDoEvent) String() (s string) {
	var word string
	if e.SerialDoType != SerialDoPendingLaunch {
		word = "at"
	} else {
		word = "since"
	}
	return fmt.Sprintf("serialDo#%s: %s#%s %s: %s",
		e.SerialDo.sdoID,
		e.SerialDoType,
		e.ID,
		word, Short(e.Time),
	)
}
