/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

// AwaitableCh is a channel whose only allowed operation is channel receive
//   - AwaitableCh implements a semaphore
//   - AwaitableCh transfers no data, instead channel close is the significant event
type AwaitableCh <-chan struct{}
