/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "fmt"

// Message is a portable routing message emitted by the netlink socket or obtain via sysctl
type Message interface {
	// Action returns portable routing message action
	Action() (action Action)
	// Dump is printable string of all values for troubleshooting
	Dump() (dump string)
	fmt.Stringer
}
