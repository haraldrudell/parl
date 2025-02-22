/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/netip"
	"slices"
	"testing"
)

func TestInterfaceAddrsMacOS(t *testing.T) {
	//t.Error("Logging on")
	var (
		macOS_lo0 = []net.Addr{&net.IPNet{
			IP:   netip.AddrFrom16(netip.MustParseAddr("127.0.0.1").As16()).AsSlice(),
			Mask: net.CIDRMask(8, 32),
		}, &net.IPNet{
			IP:   netip.MustParseAddr("::").AsSlice(),
			Mask: net.CIDRMask(128, 128),
		}, &net.IPNet{
			IP:   netip.MustParseAddr("fe80::1").AsSlice(),
			Mask: net.CIDRMask(64, 128),
		},
		}
		expIPv4 = []netip.Prefix{netip.MustParsePrefix("127.0.0.1/8")}
		expIPv6 = []netip.Prefix{
			netip.MustParsePrefix("::/128"),
			netip.MustParsePrefix("fe80::1/64"),
		}
	)

	var (
		err    error
		i4, i6 []netip.Prefix
	)

	t.Logf("macOS_lo0: %s", macOS_lo0)

	var mock = newNetInterfaceAddrsMock(macOS_lo0)
	i4, i6, err = mock.invoke()
	t.Logf("i4 %v i6 %v", i4, i6)
	if err != nil {
		t.Errorf("FAIL InterfaceAddrs err: %s", err)
	}
	if !slices.Equal(i4, expIPv4) {
		t.Errorf("FAIL i4: %v exp %v", i4, expIPv4)
	}
	if !slices.Equal(i6, expIPv6) {
		t.Errorf("FAIL i6: %v exp %v", i6, expIPv6)
	}
}

// netInterfaceAddrsMock invokes InterfaceAddrs mocked
type netInterfaceAddrsMock struct {
	fixture []net.Addr
	hook    func() (addrs []net.Addr, err error)
}

// newNetInterfaceAddrsMock returns mock invoker for InterfaceAddrs
func newNetInterfaceAddrsMock(addrs []net.Addr) (m *netInterfaceAddrsMock) {
	return &netInterfaceAddrsMock{
		fixture: addrs,
	}
}

// invoke invokes InterfaceAddrs while hook is active
func (m *netInterfaceAddrsMock) invoke() (i4, i6 []netip.Prefix, err error) {
	defer m.invokeEnd()

	m.hook = netInterfaceAddrsHook
	netInterfaceAddrsHook = m.addrsHook

	var mockIf = net.Interface{}
	i4, i6, err = InterfaceAddrs(&mockIf)

	return
}

// invokeEnd resets the hook
func (m *netInterfaceAddrsMock) invokeEnd() {
	netInterfaceAddrsHook = m.hook
	m.hook = nil
}

// addrsHook is the hook function
func (m *netInterfaceAddrsMock) addrsHook() (addrs []net.Addr, err error) {
	addrs = m.fixture

	return
}
