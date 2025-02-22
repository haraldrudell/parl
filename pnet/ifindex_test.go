/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"fmt"
	"net"
	"net/netip"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

// go test -run ^TestIfIndex_Interface$ github.com/haraldrudell/parl/pnet
func TestIfIndex_Interface(t *testing.T) {
	//t.Error("Logging on")
	const (
		// the network interface index for localhost: “lo” or “lo0”
		localInterfaceIndex = 1
		// whether to check lo interface
		examine_lo = true
	)
	var (
		// typed localhost interface index
		ifIndex = NewIfIndex(1)
		// can net.Interface be mocked?
		//	- [Interface] fields are all public, so it can be instantiated
		//	- uses [net.InterfaceByIndex]
		//	- uses [net.Interface.Addrs]
		//	- — retrieves addresses from internal structure
		// mockInterface = net.Interface{
		// 	Index: localInterfaceIndex,
		// 	Name:  "lo",
		// }
	)

	// Research: localhost interface for macOS and Linux:
	if examine_lo {
		// Interface struct {
		// 	Index        int          // positive integer that starts at one, zero is never used
		// 	MTU          int          // maximum transmission unit
		// 	Name         string       // e.g., "en0", "lo0", "eth0.100"
		// 	HardwareAddr HardwareAddr // IEEE MAC-48, EUI-48 and EUI-64 form
		// 	Flags        Flags        // e.g., FlagUp, FlagLoopback, FlagMulticast
		// }
		var _ net.Interface
		var netInterface, err = net.InterfaceByIndex(localInterfaceIndex)
		if err != nil {
			panic(err)
		}
		var addrs []net.Addr
		addrs, err = netInterface.Addrs()
		if err != nil {
			panic(err)
		}
		var s = make([]string, len(addrs))
		for i, a := range addrs {

			// string format, fails to print true IPv6 and mask: "127.0.0.1/8"
			s[i] = fmt.Sprintf("net.Addr: Network: “%s” String: “%s”",
				a.Network(), a,
			)
			if netIPNet, ok := a.(*net.IPNet); ok {
				var ip, _ = netip.AddrFromSlice(netIPNet.IP)
				var ones, bits = netIPNet.Mask.Size()
				s[i] = fmt.Sprintf("net.IP: bytes: %d “%s” mask: /%d size: %d “%s” ",
					len(netIPNet.IP), ip,
					ones, bits, netIPNet.Mask,
				) + s[i]
			} else {
				t.Errorf("net.Addr not *net.IPNet: %t", a)
			}
		}
		// GOOS: darwin interface#1: Name: "lo0" Addrs:
		// net.IP: bytes: 16 “::ffff:127.0.0.1” mask: /8 size: 32 “ff000000” net.Addr: Network: “ip+net” String: “127.0.0.1/8”
		// net.IP: bytes: 16 “::1” mask: /128 size: 128 “ffffffffffffffffffffffffffffffff” net.Addr: Network: “ip+net” String: “::1/128”
		// net.IP: bytes: 16 “fe80::1” mask: /64 size: 128 “ffffffffffffffff0000000000000000” net.Addr: Network: “ip+net” String: “fe80::1/64”
		// GOOS: linux interface#1: Name: "lo" Addrs:
		// net.IP: bytes: 16 “::ffff:127.0.0.1” mask: /8 size: 32 “ff000000” net.Addr: Network: “ip+net” String: “127.0.0.1/8”
		// net.IP: bytes: 16 “::1” mask: /128 size: 128 “ffffffffffffffffffffffffffffffff” net.Addr: Network: “ip+net” String: ::1/128”
		t.Logf("GOOS: %s interface#%d: Name: %q Addrs:\n%s",
			runtime.GOOS, localInterfaceIndex, netInterface.Name,
			strings.Join(s, "\n"),
		)
	}

	var _ net.Interface
	// uses internal structures
	var _ = net.InterfaceByIndex

	var (
		err    error
		name   string
		i4, i6 []netip.Prefix
		// netInterface *net.Interface
	)

	// Research: display net.Interface field values
	if examine_lo {
		var netInterface *net.Interface
		netInterface, err = net.InterfaceByIndex(localInterfaceIndex)
		if err != nil {
			panic(err)
		}
		// *net.Interface: &{
		// Index:1 MTU:16384 Name:lo0 HardwareAddr:
		// Flags:up|loopback|multicast|running
		// }
		t.Logf("*net.Interface: %+v", netInterface)
	}

	// interface #1 should exist
	// useNameCache missing: cache will not be used
	//	- uses [net.InterfaceByIndex]
	//	- uses [net.Interface.Addrs]
	name, i4, i6, err = ifIndex.InterfaceAddrs()
	if err != nil {
		t.Fatalf("FAIL ifIndex.Interface err: %s", perrors.Short(err))
	}

	// name should be present
	if name == "" {
		t.Error("interface#1 no name")
	}

	// net.IPnet always extracts IPv4-mapped IPv6 addresses to IPv4
	//	- mask byte-length can only be IPv4: 4 bytes or IPv6: 16 bytes
	//	- An IPv4 mask is only accepted for IPv4 address
	//	- An IPv6 mask for IPv4 address is abbreviated
	var _ = (&net.IPNet{}).String

	// all net.Addr implementations should be non-empty
	var s46List [2]string
	for i, prefixList := range [][]netip.Prefix{i4, i6} {
		var familyStrings = make([]string, len(prefixList))
		for j, prefix := range prefixList {
			var _ netip.Prefix
			var prefixString = prefix.String()
			var a = prefix.Addr()
			if a.Is6() {
				prefixString += " zone: " + strconv.Quote(a.Zone())
			}
			familyStrings[j] = prefixString
		}
		s46List[i] = fmt.Sprintf("%d[%s]", len(prefixList), strings.Join(familyStrings, "\x20"))
	}
	// name: "lo0"
	// IPv4: 0[]
	// IPv6: 3[::ffff:127.0.0.1/8 zone: "" ::1/128 zone: "" fe80::1/64 zone: ""]
	t.Logf("name: %q IPv4: %s IPv6: %s",
		name, s46List[0], s46List[1],
	)
}
