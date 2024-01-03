/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"net/netip"
	"reflect"
	"testing"
)

func TestAddrToIPAddr(t *testing.T) {
	//t.Errorf("logging on")
	var ipv6 = netip.MustParseAddr("fe80::%eth0")
	// ipv6: fe80::%eth0
	t.Logf("ipv6: %s", ipv6)
	var netIPAddr = net.IPAddr{IP: ipv6.AsSlice(), Zone: ipv6.Zone()}
	//network: "ip" string: "fe80::%eth0"
	t.Logf("network: %q string: %q", netIPAddr.Network(), &netIPAddr)

	var ipv6Socket = netip.MustParseAddrPort("[fe80::%eth0]:80")
	// ipv6Socket: [fe80::%eth0]:80
	t.Logf("ipv6Socket: %s", ipv6Socket)
	var tcp = net.TCPAddr{
		IP:   ipv6Socket.Addr().AsSlice(),
		Port: int(ipv6Socket.Port()),
		Zone: ipv6Socket.Addr().Zone(),
	}
	// tcp network: "tcp" string: "[fe80::%eth0]:80"
	t.Logf("tcp network: %q string: %q", tcp.Network(), &tcp)

	type args struct {
		addr netip.Addr
	}
	tests := []struct {
		name              string
		args              args
		wantAddrInterface net.Addr
	}{
		{"IPv6", args{ipv6}, &netIPAddr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAddrInterface := AddrToIPAddr(tt.args.addr); !reflect.DeepEqual(gotAddrInterface, tt.wantAddrInterface) {
				t.Errorf("AddrToIPAddr() = %v, want %v", gotAddrInterface, tt.wantAddrInterface)
			}
		})
	}
}
