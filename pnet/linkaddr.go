/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

const (
	HardwareAddrMac48  = 6
	HardwareAddrEui64  = 8
	HardwareAddrInfini = 20
)

// ErrNoSuchInterface is the value net.errNoSuchInterface
//
// Usage:
//
//	if errors.Is(err, pnet.ErrNoSuchInterface) { …
var ErrNoSuchInterface = func() (err error) {
	for _, e := net.InterfaceByName("a b"); e != nil; e = errors.Unwrap(e) {
		err = e
	}
	if err == nil {
		panic(perrors.NewPF("failed to obtain NoSuchInterface from InterfaceByName"))
	}
	return
}()

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
func NewLinkAddr(index IfIndex, name string) (linkAddr *LinkAddr) {
	return &LinkAddr{
		IfIndex: index,
		Name:    name,
	}
}

func (a *LinkAddr) UpdateFrom(b *LinkAddr) (isComplete bool) {
	if !a.IfIndex.IsValid() && b.IfIndex.IsValid() {
		a.IfIndex = b.IfIndex
	}
	if a.Name == "" && b.Name != "" {
		a.Name = b.Name
	}
	if len(a.HardwareAddr) == 0 && len(b.HardwareAddr) > 0 {
		a.HardwareAddr = b.HardwareAddr
	}
	return a.IsComplete()
}

func (a *LinkAddr) SetHw(hw net.HardwareAddr) (err error) {
	if !slices.Contains(HardwareAddrLengthsWithZero, len(hw)) {
		err = perrors.ErrorfPF("hardware address bad length: %d allowed: [%v]", hw)
		return
	} else if len(hw) > 0 {
		a.HardwareAddr = hw
	}
	return
}

func (a *LinkAddr) SetName(name string) {
	a.Name = name
}

// UpdateName attempts to populate interface name if not already present
func (a *LinkAddr) UpdateName() (linkAddr *LinkAddr, err error) {
	linkAddr = a
	if a.Name != "" {
		return // name already present return
	}
	var name string
	if name, _, err = a.IfIndex.Zone(); err != nil {
		return // error while getting interface data return
	}
	if name == "" {
		return // no new name obtained return
	}
	linkAddr = &LinkAddr{
		IfIndex:      a.IfIndex,
		Name:         name,
		HardwareAddr: a.HardwareAddr,
	}
	return // name updated return
}

// Interface returns net.Interface associated with LinkAddr
//   - order is index, name, mac
//   - if LinkAddr is zero-value, nil is returned
func (a *LinkAddr) Interface() (netInterface *net.Interface, isNoSuchInterface bool, err error) {
	if a.IfIndex.IsValid() {
		return a.IfIndex.Interface()
	} else if a.Name != "" {
		if netInterface, err = net.InterfaceByName(a.Name); err != nil {
			isNoSuchInterface = errors.Is(err, ErrNoSuchInterface)
			err = perrors.Errorf("net.InterfaceByName %w", err)
		}
		return
	} else if len(a.HardwareAddr) > 0 {
		return HardwareAddrInterface(a.HardwareAddr)
	}
	return // zero-value: netInterface nil return
}

// ZoneID is the IPv6 ZoneID for this interface
func (a *LinkAddr) ZoneID() string {
	if a != nil {
		if a.Name != "" {
			return a.Name
		} else if a.IfIndex > 0 {
			return strconv.Itoa(int(a.IfIndex))
		}
	}
	return "0"
}

// OneString picks the most meaningful value
//   - interface name or hardware address or #interface index or "0"
func (a *LinkAddr) OneString() string {
	if a != nil {
		if a.Name != "" {
			return a.Name
		} else if len(a.HardwareAddr) > 0 {
			return a.HardwareAddr.String()
		} else if a.IfIndex > 0 {
			return "#" + strconv.Itoa(int(a.IfIndex))
		}
	}
	return "0"
}

func (a *LinkAddr) IsValid() (isValid bool) {
	return a.IfIndex != 0 ||
		a.Name != "" ||
		len(a.HardwareAddr) > 0
}

func (a *LinkAddr) IsZeroValue() (isZeroValue bool) {
	return !a.IfIndex.IsValid() &&
		a.Name == "" &&
		a.HardwareAddr == nil
}

func (a *LinkAddr) IsComplete() (isComplete bool) {
	return a.IfIndex.IsValid() &&
		a.Name != "" &&
		len(a.HardwareAddr) > 0
}

// "#13_en5_00:00:5e:00:53:01"
//   - zero-values are skipped
//   - zero-value: "zero-value"
func (a *LinkAddr) NonZero() (s string) {
	var sL []string
	if a.IfIndex != 0 {
		sL = append(sL, a.IfIndex.String()) // "#13"
	}
	if a.Name != "" {
		sL = append(sL, a.Name)
	}
	if len(a.HardwareAddr) > 0 {
		sL = append(sL, a.HardwareAddr.String()) // "00:00:5e:00:53:01"
	}
	if len(sL) > 0 {
		s = strings.Join(sL, "_")
	} else {
		s = "zero-value"
	}
	return
}

func (a *LinkAddr) Dump() (s string) {
	return parl.Sprintf("linkAddr#%d%q_hw%s",
		a.IfIndex,
		a.Name,
		a.HardwareAddr.String(),
	)
}

// "en8(28)00:11:22:33:44:55:66"
func (a *LinkAddr) String() (s string) {
	if len(a.Name) > 0 {
		s += a.Name
	}
	if a.IfIndex > 0 {
		s += "(" + strconv.Itoa(int(a.IfIndex)) + ")"
	}
	s += a.HardwareAddr.String()
	return
}
