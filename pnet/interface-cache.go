/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

const (
	NoCache  NameCacher = iota // name cache should not be used in the api call
	Update                     // use name cache and update the cache first
	NoUpdate                   // use name cache without update. Mechanic for ongoing name cache update is required
)

// NameCacher contaisn instructions for interface-name cache read: [NoCache] [Update] [NoUpdate]
type NameCacher uint8

// InterfaceCache is a cache mapping per-boot stable network interface index #1 to name "lo0". Thread-Safe
type InterfaceCache struct {
	updateLock sync.Mutex
	m          atomic.Pointer[map[IfIndex]string] // write behind lock
}

// NewInterfaceCache returns a cache mapiing network-interfac eindex to name. Thread-Safe
func NewInterfaceCache() (interfaceCache *InterfaceCache) {
	return &InterfaceCache{}
}

// Init loads the cache, an opeation that may fail. Funtional chaining. Thread-Safe
func (i *InterfaceCache) Init() (i0 *InterfaceCache) {
	if _, err := i.Update(); err != nil {
		panic(err)
	}
	return i
}

// CachedName retrieves a name by index with optional cache update. Thread-Safe
//   - default is to update
//   - unknown index returns empty string
func (i *InterfaceCache) CachedName(ifIndex IfIndex, noUpdate ...NameCacher) (name string, err error) {
	var cacher = Update
	if len(noUpdate) > 0 {
		cacher = noUpdate[0]
	}

	if cacher != Update {
		name = i.CachedNameNoUpdate(ifIndex)
		return
	}

	var m map[IfIndex]string
	if m, err = i.Update(); err != nil {
		return
	}
	name = m[ifIndex]

	return
}

// CachedNameNoUpdate retrieves a name by index with no cache update. Thread-Safe
func (i *InterfaceCache) CachedNameNoUpdate(ifIndex IfIndex) (name string) {
	var m map[IfIndex]string
	if mp := i.m.Load(); mp != nil {
		m = *mp
	} else {
		panic(perrors.NewPF("map never updated"))
	}

	name = m[ifIndex]
	return
}

// Update updates the cache by reading all system interfaces. Thread-Safe
func (i *InterfaceCache) Update() (m map[IfIndex]string, err error) {
	i.updateLock.Lock()
	defer i.updateLock.Unlock()

	// get interfaces
	var interfaces []net.Interface
	if interfaces, err = Interfaces(); err != nil {
		return
	}

	// get map
	if mp := i.m.Load(); mp != nil {
		m = *mp
	} else {
		m = make(map[IfIndex]string)
	}

	// populate map
	for ix := 0; ix < len(interfaces); ix++ {
		ifp := &interfaces[ix]
		if ifp.Name == "" {
			continue
		}
		var ifIndex IfIndex
		if ifIndex, err = NewIfIndexInt(ifp.Index); err != nil {
			return
		}
		m[ifIndex] = ifp.Name
	}

	// store atomically
	i.m.Store(&m)

	return
}

// Map returns a copy of the current cache. Thread-Safe
func (i *InterfaceCache) Map() (m map[IfIndex]string) {
	i.updateLock.Lock()
	defer i.updateLock.Unlock()

	if mp := i.m.Load(); mp != nil {
		m = make(map[IfIndex]string, len(*mp))
		for k, v := range *mp {
			m[k] = v
		}
	}
	return
}

// Map replaces the cache. Thread-Safe
//   - do not read from or write to newMap after SetMap
func (i *InterfaceCache) SetMap(newMap map[IfIndex]string) (oldMap map[IfIndex]string) {
	i.updateLock.Lock()
	defer i.updateLock.Unlock()

	if mp := i.m.Swap(&newMap); mp != nil {
		oldMap = *mp
	}
	return
}
