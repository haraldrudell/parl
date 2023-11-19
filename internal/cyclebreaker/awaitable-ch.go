/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package cyclebreaker

// AwaitableCh is a one-to-many inter-thread wait-mechanic with happens-before
//   - AwaitableCh implements a semaphore
//   - implementation is a channel whose only allowed operation is channel receive
//   - AwaitableCh transfers no data, instead channel close is the significant event
//
// Usage:
//
//	<-ch // waits for event
//
//	select {
//	  case <-ch:
//	    hasHappened = true
//	  default:
//	    hasHappened = false
//	}
type AwaitableCh <-chan struct{}
