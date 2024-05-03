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
	//	- may only be invoked once or panic
	//	- the provided connection is guaranteed to be closed
	//	- thread-safe
	//	- Handle invocations after Shutdown immediately
	//		close the connection
	Handle(conn C)
	// may be invoked at any time: before, after or in lieu of Handle
	//	- on Shutdown return, any provided connection has been
	//		closed and all resources have been released
	//	- thread-safe idempotent
	Shutdown()
	fmt.Stringer
}
