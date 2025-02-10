/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// InterfaceCache is a cache mapping per-boot stable network interface index #1 to name "lo0". Thread-Safe
type InterfaceCache struct {
	// makes m thread-safe
	updateLock parl.Mutex
	// map behind lock
	m map[IfIndex]string
}

// NewInterfaceCache returns a cache mapping network-interface index to name
func NewInterfaceCache() (interfaceCache *InterfaceCache) { return &InterfaceCache{} }

// Init loads the cache, an operation that may fail. Funtional chaining. Thread-Safe
//   - used for initializing a static-variable cache
func (i *InterfaceCache) Init() (i0 *InterfaceCache) {
	if _, err := i.CachedName(); err != nil {
		panic(err)
	}
	return i
}

// CachedName retrieves a name by index after cache update
//   - ifIndex: optional query to return in name after cache update
//   - unknown index returns empty string
//   - Thread-Safe
func (i *InterfaceCache) CachedName(ifIndex ...IfIndex) (name string, err error) {
	defer i.updateLock.Lock().Unlock()

	// get interfaces
	var interfaces []net.Interface
	if interfaces, err = Interfaces(); err != nil {
		return
	}

	// get map to use
	if i.m == nil {
		i.m = make(map[IfIndex]string, len(interfaces))
	}

	// populate map
	for ix := 0; ix < len(interfaces); ix++ {
		var interface1 = &interfaces[ix]
		var name1 = interface1.Name
		var ifIndex1 IfIndex
		if name == "" {
			continue // nameless interface ignore
		} else if ifIndex1, err = NewIfIndexInt(interface1.Index); err != nil {
			return // interface without index error
		} else if i.m[ifIndex1] != name1 {
			// write only updated values
			i.m[ifIndex1] = name1
		}
	}

	// carry out any provided query
	if len(ifIndex) == 0 {
		return // no provided query
	}
	name = i.m[ifIndex[0]]

	return
}

// CachedNameNoUpdate retrieves a name by index with no cache update. Thread-Safe
func (i *InterfaceCache) CachedNameNoUpdate(ifIndex IfIndex) (name string) {
	defer i.updateLock.Lock().Unlock()

	if i.m == nil {
		panic(perrors.NewPF("map never updated"))
	}
	name = i.m[ifIndex]

	return
}

// Map replaces the cache. Thread-Safe
//   - do not read from or write to newMap after SetMap
//   - used for testing
func (i *InterfaceCache) SetMap(newMap map[IfIndex]string) (oldMap map[IfIndex]string) {
	defer i.updateLock.Lock().Unlock()

	oldMap = i.m
	i.m = newMap

	return
}
