/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "strconv"

const (
	// The SerialDo is invoking thunk from idle, now time
	SerialDoLaunch SerialDoType = 1 + iota
	// The SerialDo enqueued a future invocation, request time
	SerialDoPending
	// The SerialDo is launching a pending invocation, request time
	SerialDoPendingLaunch
	// The SerialDo returned to being idle, time is busy since
	SerialDoIdle
)

var serialDoEventToString = map[SerialDoType]string{
	SerialDoLaunch:        "launch",
	SerialDoPending:       "enqueue",
	SerialDoPendingLaunch: "launch-enqueued",
	SerialDoIdle:          "idle",
}

type SerialDoType uint8

func (e SerialDoType) IsValid() (isValid bool) {
	_, isValid = serialDoEventToString[e]
	return
}

func (e SerialDoType) String() (s string) {
	var ok bool
	if s, ok = serialDoEventToString[e]; !ok {
		s = "unknown: " + strconv.FormatUint(uint64(e), 10)
	}
	return
}
