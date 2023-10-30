/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import "fmt"

const (
	// iteration is in progress
	notCanceled cancelStates = iota
	// consumer has invoked Cancel
	cancelRequested
	// cancel was successfully requested from the iterator-value
	// producer
	cancelComplete
	// end-of-data notice received from value-producer,
	// typically by returning parl.ErrEndCallbacks
	endOfData
	// value-producer returned error other than parl.ErrEndCallbacks
	errorReceived
)

// notCanceled cancelRequested cancelComplete endOfData errorReceived
type cancelStates uint32

var cancelStatesMap = map[cancelStates]string{
	notCanceled:     "notCanceled",
	cancelRequested: "cancelRequested",
	cancelComplete:  "cancelComplete",
	endOfData:       "endOfData",
	errorReceived:   "errorReceived",
}

func (s cancelStates) String() (s2 string) {
	s2 = cancelStatesMap[s]
	if s2 == "" {
		s2 = fmt.Sprintf("?badCancelState:%d", s)
	}
	return
}
