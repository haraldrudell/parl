/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlnet

import (
	"fmt"
	"net"
	"strconv"
)

// LinkAddr contains an Ethernet mac address, its interface name and interface index
type LinkAddr struct {
	Index IfIndex          // 0 is none
	Name  string           // "" none
	Hw    net.HardwareAddr // []byte
}

// NewLinkAddr instantiates LinkAddr
func NewLinkAddr(index IfIndex) *LinkAddr {
	return &LinkAddr{Index: index}
}

// NewLinkAddr2 instantiates LinkAddr
func NewLinkAddr2(index IfIndex, name string, hw net.HardwareAddr) *LinkAddr {
	return &LinkAddr{Index: index, Name: name, Hw: hw}
}

// UpdateName attempts to populate interface name if not already present
func (a *LinkAddr) UpdateName() (err error) {
	if a.Name == "" {
		a.Name, err = a.Index.Zone()
	}
	return
}

// ZoneID is the IPv6 ZoneID for this interface
func (a *LinkAddr) ZoneID() string {
	if a != nil {
		if a.Name != "" {
			return a.Name
		} else if a.Index > 0 {
			return strconv.Itoa(int(a.Index))
		}
	}
	return "0"
}

// OneString picks the most meaningful value
func (a *LinkAddr) OneString() string {
	if a != nil {
		if a.Name != "" {
			return a.Name
		} else if len(a.Hw) > 0 {
			return a.Hw.String()
		} else if a.Index > 0 {
			return strconv.Itoa(int(a.Index))
		}
	}
	return "0"
}

func (a LinkAddr) String() string {
	return fmt.Sprintf("%s(%d)%s", a.Name, a.Index, a.Hw)
}
