/*
© 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// ChannelReceive is deferrable function that waits for a channel
//   - ch: a channel to receive from
//   - — may be nil
//   - — may be closed
//   - — may block due to being unbuffered or empty
//   - —
//   - deferrable
func ChannelReceive[T any](ch <-chan T) {
	NilPanic("channel", ch)
	if _ /*value*/, _ /*hasValue*/, err := channelReceive(ch); err != nil {
		// channel was closed
		panic(err)
	}
}

func channelReceive[T any](ch <-chan T) (value T, hasValue bool, err error) {
	defer PanicToErr(&err)

	value, hasValue = <-ch

	return
}
