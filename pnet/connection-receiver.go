/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net"
)

// ConnectionReceiver is used by [ThreadSource] to
// process or cancel connections
//   - ThreadSource provides ahead notice so that any
//     thread-strategy can be implemented
//   - either Handle or Shutdown or both must be invoked
//     for all obtained ConnectionReceivers
type ConnectionReceiver[C net.Conn] interface {
	// Handle operates on and closes a connection
	//	- Handle may be invoked zero or one time
	//	- If Handle is not invoked, Shutdown is invoked
	//	- conn is never nil and must be closed prior to Handle return
	//	- a Handle invocation after Shutdown must immediately
	//		close conn and return
	//	- a Shutdown invocation during Handle means that Handle should return
	//		as soon as practical: the consumer is waiting for its return
	//	- Handle must be thread-safe versus Shutdown and [ThreadSource.Receiver]
	Handle(conn C)
	// Shutdown may be invoked any time and more than once: before, after or in lieu of Handle
	//	- once Shutdown and any pending Handle invocation have returned:
	//	- — resources must have been released and any connection closed
	//	- Shutdown must be thread-safe and idempotent
	Shutdown()
	fmt.Stringer
}
