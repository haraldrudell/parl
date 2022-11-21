/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

const (
	HardwareAddrMac48  = 6
	HardwareAddrEui64  = 8
	HardwareAddrInfini = 20
)

var HardwareAddrLengths = []int{HardwareAddrMac48, HardwareAddrEui64, HardwareAddrInfini}
var HardwareAddrLengthsWithZero = append([]int{0}, HardwareAddrLengths...)

func IsHardwareAddrLength(byts []byte) (isHardwareAddrLength bool) {
	return slices.Contains(HardwareAddrLengths, len(byts))
}

// LinkAddr contains an Ethernet mac address, its interface name and interface index
type LinkAddr struct {
	IfIndex                 // 0 is none
	Name             string // "" none
	net.HardwareAddr        // []byte
}

// NewLinkAddr instantiates LinkAddr
func NewLinkAddr(index IfIndex, name string, hw net.HardwareAddr) (linkAddr *LinkAddr, err error) {

	// check hw
	var hardwareAddr net.HardwareAddr
	if !slices.Contains(HardwareAddrLengthsWithZero, len(hw)) {
		err = perrors.ErrorfPF("hardware address bad length: %d allowed: [%v]", hardwareAddr)
		return
	} else if len(hw) > 0 {
		hardwareAddr = hw
	}

	linkAddr = &LinkAddr{
		IfIndex:      index,
		Name:         name,
		HardwareAddr: hardwareAddr,
	}

	return
}

// UpdateName attempts to populate interface name if not already present
func (linkA *LinkAddr) UpdateName() (linkAddr *LinkAddr, err error) {
	linkAddr = linkA
	if linkA.Name != "" {
		return // name already present return
	}
	var name string
	if name, err = linkA.IfIndex.Zone(); err != nil {
		return // error while getting interface data return
	}
	if name == "" {
		return // no new name obtained return
	}
	linkAddr = &LinkAddr{
		IfIndex:      linkA.IfIndex,
		Name:         name,
		HardwareAddr: linkA.HardwareAddr,
	}
	return // name updated return
}

// ZoneID is the IPv6 ZoneID for this interface
func (linkA *LinkAddr) ZoneID() string {
	if linkA != nil {
		if linkA.Name != "" {
			return linkA.Name
		} else if linkA.IfIndex > 0 {
			return strconv.Itoa(int(linkA.IfIndex))
		}
	}
	return "0"
}

// OneString picks the most meaningful value
func (linkA *LinkAddr) OneString() string {
	if linkA != nil {
		if linkA.Name != "" {
			return linkA.Name
		} else if len(linkA.HardwareAddr) > 0 {
			return linkA.HardwareAddr.String()
		} else if linkA.IfIndex > 0 {
			return "#" + strconv.Itoa(int(linkA.IfIndex))
		}
	}
	return "0"
}

// "en8(28)00:11:22:33:44:55:66"
func (linkA *LinkAddr) String() (s string) {
	if len(linkA.Name) > 0 {
		s += linkA.Name
	}
	if linkA.IfIndex > 0 {
		s += "(" + strconv.Itoa(int(linkA.IfIndex)) + ")"
	}
	s += linkA.HardwareAddr.String()
	return
}
