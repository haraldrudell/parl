/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

var networkInterfaceNameCache = NewInterfaceCache().Init()

func UpdateNameCache() (err error) {
	return networkInterfaceNameCache.Update()
}

func CachedName(ifIndex IfIndex) (name string) {
	return networkInterfaceNameCache.CachedNameNoUpdate(ifIndex)
}

func NameCache() (m map[IfIndex]string) {
	return networkInterfaceNameCache.Map()
}
