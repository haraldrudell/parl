/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

// UnpackEvents extracts events and slices from a value
// received on [WatcherCh.Ch]
//   - interface type is any
//   - runtime type is *WatchEvent or []*WatchEvent
//   - if watchEvent non-nil, it is the single event returned
//   - if len(watchEvents) > 0, it holds 2 or more returned events
//   - if watchEvent and watchEvents are both nil, the watcher has ended
func UnpackEvents(anyEvent any) (watchEvent *WatchEvent, watchEvents []*WatchEvent) {

	// ended case
	if anyEvent == nil {
		return // ended return
	}

	// single event case
	var ok bool
	if watchEvent, ok = anyEvent.(*WatchEvent); ok {
		return // single event return
	}
	// multiple events case
	watchEvents = anyEvent.([]*WatchEvent)

	return // multiple events return
}
