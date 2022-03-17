/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlnet

import "fmt"

// Message is a portable routing message emitted by netlink socket
type Message interface {
	fmt.Stringer
}

// Callback allows for processing of routing message,eg. populating a map
type Callback func(msg Message)
