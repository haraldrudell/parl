/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
)

// LazyCyclic is CyclicAwaitable initialized on first use
type LazyCyclic struct {
	// if false, the cyclic is not in active use
	//	- the cyclic starts in Open state
	//	- the consumer sets IsActive to true once the cyclic’s initial state is established
	//	- CompareAndSwap can be used for selecting winner initializing thread in
	//		eventually consistent designs
	//	- IsActive can shield locks with atomic performance prior to
	//		the LazyCyclic being provided to other threads or deteremined to be active
	IsActive atomic.Bool
	// Lock atomizes operations [LazyCyclic.Cyclic.Open] and [LazyCyclic.Cyclic.Close]
	// with its justifying observations
	Lock Mutex
	// Cyclic contains a closing channel that can be re-opened
	Cyclic CyclicAwaitable
}
