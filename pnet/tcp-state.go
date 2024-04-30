/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

const (
	// the socket is listening from a successful invocation of [SocketListener.Listen]
	soListening socketState = iota + 1
	soAccepting
	soClosing
	soClosed
)

// socketState represents the state of a tcp udp unix ip socket listener
//   - soListening soAccepting soClosing soClosed
//   - the socket starts with value 0: not listening
type socketState uint32
