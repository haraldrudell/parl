/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "fmt"

// Message is a portable routing message emitted by the netlink socket or obtain via sysctl
type Message interface {
	Action() (action Action)
	fmt.Stringer
}

// Callback allows for processing of routing message,eg. populating a map
type Callback func(msg Message)
