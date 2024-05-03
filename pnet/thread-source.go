/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"

	"github.com/haraldrudell/parl"
)

// ThreadSource provides a thread-allocation strategy
//   - when Receiver is invoked, threads can be pre-launched for
//     a pending connection
//   - default allocation strategy when ThreadSource is not used is
//     one thread per connection
//   - ThreadSource allows for any thread-allocation strategy to be used
//     when handling multiple connections
//   - when connReceiver.Handle is invoked, threads launched by Receiver
//     may already be waiting at a a lock for maximum performance
type ThreadSource[C net.Conn] interface {
	// Receiver prepares the ThreadSource for a possible upcoming connection
	//	- done will be invoked exactly once by all connReceiver objects
	//	- to ensure done invocation, [ConnectionReceiver.Handle] or
	//		[ConnectionReceiver.Shutdown] or both must be invoked
	//	- Receiver and connReceiver.Handle do not block
	//	- any connection provided to [ConnectionReceiver.Handle] is
	//		guaranteed to be closed by Handle even if
	//		[ConnectionReceiver.Shutdown] was already invoked
	Receiver(done func(), addError parl.AddError) (connReceiver ConnectionReceiver[C], err error)
}
