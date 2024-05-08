/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/sets"

const (
	// NBChanAlways configures NBChan to always have a thread
	//   - benefit: for empty NBChan, Send/SendMany to  channel receive is faster due to avoiding thread launch
	//   - cost: a thread is always running instad of only running when NBChan non-empty
	//   - cost: Close or CloseNow must be invoked to shutdown NBChan
	NBChanAlways NBChanThreadType = iota + 1
	// NBChanNone configures no thread.
	//	- benefit: lower cpu
	//	- cost: data can only be received using [NBChan.Get]
	NBChanNone
)

// NBChanThreadType defines how NBChan is operating
//   - [NBChanAlways] [NBChanNone] or on-demand thread
type NBChanThreadType uint8

func (n NBChanThreadType) String() (s string) {
	return nbChanThreadTypeSet.StringT(n)
}

// set for [NBChanThreadType]
var nbChanThreadTypeSet = sets.NewSet[NBChanThreadType]([]sets.SetElement[NBChanThreadType]{
	{ValueV: NBChanAlways, Name: "alwaysCh"},
	{ValueV: NBChanNone, Name: "noCh"},
})
