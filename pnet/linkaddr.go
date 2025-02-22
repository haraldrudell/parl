/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// byte length of 48-bit hardware address
	HardwareAddrMac48 = 6
	// byte length of 64-bit hardware address
	HardwareAddrEui64 = 8
	// byte length of 160-bit hardware address
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

// hardwareAddrLengthsMap is O(1) allowed harware address lengths
var hardwareAddrLengthsMap = map[int]struct{}{
	HardwareAddrMac48:  {},
	HardwareAddrEui64:  {},
	HardwareAddrInfini: {},
}

// “6 8 20”
var goodList = fmt.Sprintf("%d %d %d",
	HardwareAddrMac48, HardwareAddrEui64, HardwareAddrInfini,
)

const (
	// [IsHardwareAddrLength] zeroOK: zero-length is OK
	HwZeroOK HWZero = 1
)

// [HwZeroOK]: [IsHardwareAddrLength] zeroOK: zero-length is OK
type HWZero uint8

// IsHardwareAddrLength returns err: nil if
// byte-slice length is an allowed hardware address length
//   - byts candidate bytes for hardware address
//   - — [net.HardwareAddr] implementation is []byte
//   - zeroOK true: zero length is OK, too 0, 6, 8 20 bytes
//   - zeroOk false: 6, 8, 20 bytes
func IsHardwareAddrLength(mac []byte, zeroOK ...HWZero) (err error) {
	var isZeroOK = len(zeroOK) > 0 && zeroOK[0] == HwZeroOK

	if isZeroOK && len(mac) == 0 {
		return // zero-length OK return
	} else if _, ok := hardwareAddrLengthsMap[len(mac)]; ok {
		return // good length return
	}

	// bad return
	var lengths = goodList
	if isZeroOK {
		lengths = "0 " + lengths
	}
	err = perrors.ErrorfPF("bad mac net.HardwareAddr length: %d not: %s",
		len(mac), lengths,
	)

	return
}

// LinkAddr contains an Ethernet mac address, its interface name and interface index
type LinkAddr struct {
	// 0 is none
	//	- a host numbers network interfaces 1… guaranteed stable until next reboot
	//	- operating systems tries to make index stable across reboots
	//	- 1 is typically the local network interface
	IfIndex
	// empty for none, “lo0” “eth0” “lo”
	Name string
	// []byte physical hardware address
	//	- 6, 8 or 20 bytes
	//	- colon, hyphen or period separators between digits
	//	- grouped 2-digit lower-case hex. For period separator, 4-digit
	net.HardwareAddr
}

// NewLinkAddr instantiates LinkAddr
func NewLinkAddr(index IfIndex, name string) (linkAddr *LinkAddr) {
	return &LinkAddr{
		IfIndex: index,
		Name:    name,
	}
}

// UpdateFrom copies any values in b that are not in a
//   - returns whether all fields in a are now initialized
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

// SetHw sets hardware address, zero-length allowed
func (a *LinkAddr) SetHw(hw net.HardwareAddr) (err error) {
	if err = IsHardwareAddrLength(hw, HwZeroOK); err != nil {
		return
	} else if len(hw) > 0 {
		a.HardwareAddr = hw
	}

	return
}

// SetName updates network interface name
func (a *LinkAddr) SetName(name string) { a.Name = name }

// UpdateName attempts to populate interface name if not already present
func (a *LinkAddr) UpdateName() (linkAddr *LinkAddr, err error) {

	// check if network interface name is already present
	linkAddr = a
	if a.Name != "" {
		return // name already present return
	}

	// get network interface name from possible interface index string-zone
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
//   - search field order is index, name, mac
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
//   - if non-empty network interface name present
//   - if non-zero network interface index numeric string
//   - otherwise “0”
//   - never empty
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
//   - interface name or hardware address or #interface index or “0”
//   - never empty
func (a *LinkAddr) OneString() (s string) {
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

// IsValid returns true if at least one field of LinkAddr has been set
func (a *LinkAddr) IsValid() (isValid bool) {
	return a.IfIndex != 0 ||
		a.Name != "" ||
		len(a.HardwareAddr) > 0
}

// IsZeroValue returns true if LinkAddr is uninitialized zero-value
func (a *LinkAddr) IsZeroValue() (isZeroValue bool) {
	return !a.IfIndex.IsValid() &&
		a.Name == "" &&
		a.HardwareAddr == nil
}

// IsComplete returns true if all fields of LinkAddr have been initialized
func (a *LinkAddr) IsComplete() (isComplete bool) {
	return a.IfIndex.IsValid() &&
		a.Name != "" &&
		len(a.HardwareAddr) > 0
}

// “#13_en5_00:00:5e:00:53:01” “eth0”
//   - zero-values are skipped
//   - all zero-values: “zero-value”
//   - never empty string
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

// Dump retuns all fields of LinkAddr for troubleshooting
//   - “linkAddr#28"en8"_hw00:11:22:33:44:55:66”
//   - zero-value: “linkAddr#0""_hw”
func (a *LinkAddr) Dump() (s string) {
	return parl.Sprintf("linkAddr#%d%q_hw%s",
		a.IfIndex,
		a.Name,
		a.HardwareAddr.String(),
	)
}

// “en8(28)00:11:22:33:44:55:66”
//   - for zero-value: empty string
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
