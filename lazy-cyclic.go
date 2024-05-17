/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
)

type LazyCyclic struct {
	// if false, the cyclic is not in active use
	IsActive atomic.Bool
	// Lock atomizes operations Cyclic.Open and Cyclic.Close
	// with its justifying observations
	Lock sync.Mutex
	// Cyclic contains a closing channel
	Cyclic CyclicAwaitable
}
