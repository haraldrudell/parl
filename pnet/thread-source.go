/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"

	"github.com/haraldrudell/parl"
)

// ThreadSource allows for a thread-allocation strategy
//   - for example a listener receives incoming connections for processing.
//     ThreadSource allows for those connections to be handed to a function
//     obtained ahead of time, thus permitting for a
//     thread-allocation strategy to be implemented
//   - when ThreadSource is not used, the default allocation strategy is
//     one new thread for each connection
//   - when Receiver is invoked, threads can created ahead of time for
//     future connections and a connection-queue can be implemented
//   - once an incoming connection occurs, connReceiver.Handle is invoked,
//     and that function may use threads already created by the ThreadSource
//     held waiting at a lock
//   - the ThreadSource does not have shutdown. Once the listeners have concluded
//     the ThreadSource can be shut down
//   - ConnectionReceiver has Shutdown and a strict protocol to prevent
//     resource leaks
type ThreadSource[C net.Conn] interface {
	// Receiver prepares the ThreadSource for a possible upcoming connection
	//	- done.Done must be invoked exactly once for all connReceiver objects
	//	- — to ensure done invocation:
	//	- — [ConnectionReceiver.Handle] or
	//	- — [ConnectionReceiver.Shutdown] or both may be invoked
	//	- Receiver and connReceiver.Handle must not block
	//	- any connection provided to [ConnectionReceiver.Handle] must be
	//		closed by Handle even if [ConnectionReceiver.Shutdown] was already invoked
	Receiver(done parl.Done, errorSink parl.ErrorSink1) (connReceiver ConnectionReceiver[C], err error)
}
