/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import "sync/atomic"

var networkInterfaceNameCache = NewInterfaceCache().Init()
var updateDisabled atomic.Bool

func UpdateNameCache() (err error) {
	if !updateDisabled.Load() {
		_, err = networkInterfaceNameCache.Update()
	}
	return
}

func CachedName(ifIndex IfIndex) (name string) {
	return networkInterfaceNameCache.CachedNameNoUpdate(ifIndex)
}

func NameCache() (m map[IfIndex]string) {
	return networkInterfaceNameCache.Map()
}

func SetTestMap(testMap map[IfIndex]string, disableUpdate bool) (oldMap map[IfIndex]string) {
	updateDisabled.Store(disableUpdate)
	return networkInterfaceNameCache.SetMap(testMap)
}
