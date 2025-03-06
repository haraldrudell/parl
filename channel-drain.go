/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

const (
	// [CloseChannel]: drain channel prior to closing it
	DoDrain = true
)

// optional argument to [CloseChannel]: [DoDrain]
type ChannelDrain bool
