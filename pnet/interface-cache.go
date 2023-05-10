/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"sync"
)

const (
	NoCache nameCacher = iota
	Update
	NoUpdate
)

type nameCacher uint8

type InterfaceCache struct {
	updateLock sync.Mutex
	m          map[IfIndex]string
}

func NewInterfaceCache() (interfaceCache *InterfaceCache) {
	return &InterfaceCache{m: make(map[IfIndex]string)}
}

func (i *InterfaceCache) Init() (i2 *InterfaceCache) {
	if err := i.Update(); err != nil {
		panic(err)
	}
	return i
}

func (i *InterfaceCache) CachedName(ifIndex IfIndex, noUpdate ...nameCacher) (name string, err error) {
	var cacher = Update
	if len(noUpdate) > 0 {
		cacher = noUpdate[0]
	}

	if cacher == Update {
		if err = i.Update(); err != nil {
			return
		}
	}
	name = i.CachedNameNoUpdate(ifIndex)
	return
}

func (i *InterfaceCache) CachedNameNoUpdate(ifIndex IfIndex) (name string) {
	return i.m[ifIndex]
}

func (i *InterfaceCache) Update() (err error) {
	i.updateLock.Lock()
	defer i.updateLock.Unlock()

	var interfaces []net.Interface
	if interfaces, err = Interfaces(); err != nil {
		return
	}
	for ix := 0; ix < len(interfaces); ix++ {
		ifp := &interfaces[ix]
		var ifIndex IfIndex
		if ifIndex, err = NewIfIndexInt(ifp.Index); err != nil {
			return
		}
		i.m[ifIndex] = ifp.Name
	}
	return
}

func (i *InterfaceCache) Map() (m map[IfIndex]string) {
	return i.m
}
