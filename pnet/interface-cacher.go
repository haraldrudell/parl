/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "sync/atomic"

// networkInterfaceNameCache is a static cache remembering interface names
//   - the cache is used by pnet functions and may be occasionally updated
var networkInterfaceNameCache = NewInterfaceCache().Init()

// test hook
var hookUpdateDisabled atomic.Bool

// UpdateNameCache updates the static cache for local network interface names
//   - provides names of interfaces that are no longer up
//   - allows consumers to force cache update after starting an interface that may be set to down later
//   - the map is updated on executable launch
//   - Thread-Safe
func UpdateNameCache() (err error) {
	if !hookUpdateDisabled.Load() {
		// cached name no query: do update only
		_, err = networkInterfaceNameCache.CachedName()
	}
	return
}

// CachedName returns a possible name for an index. Thread-Safe
//   - allows consumers to retrieve names of interfaces no longer up
func CachedName(ifIndex IfIndex) (name string) {
	return networkInterfaceNameCache.CachedNameNoUpdate(ifIndex)
}

// setTestMap sets a mapping and disables update
//   - used for testing
func SetTestMap(testMap map[IfIndex]string, disableUpdate bool) (oldMap map[IfIndex]string) {
	hookUpdateDisabled.Store(disableUpdate)
	return networkInterfaceNameCache.SetMap(testMap)
}
