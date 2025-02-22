/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net/netip"
	"testing"
)

func TestIsDirect(t *testing.T) {
	const (
		isDirectNo  = false
		isDirectYes = true
	)
	var (
		invalidPrefix netip.Prefix
	)
	type args struct {
		route netip.Prefix
	}
	tests := []struct {
		name         string
		args         args
		wantIsDirect bool
	}{
		{"invalid", args{invalidPrefix}, isDirectNo},
		{"IPv4 /0", args{DefaultRouteIPv4}, isDirectNo},
		{"IPv4 /32", args{netip.MustParsePrefix("0.0.0.0/32")}, isDirectYes},
		{"IPv6 /0", args{DefaultRouteIPv6}, isDirectNo},
		{"IPv6 /128", args{netip.MustParsePrefix("::/128")}, isDirectYes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsDirect := IsDirect(tt.args.route); gotIsDirect != tt.wantIsDirect {
				t.Errorf("IsDirect() = %v, want %v", gotIsDirect, tt.wantIsDirect)
			}
		})
	}
}

func TestIsBroadcast(t *testing.T) {
	//t.Error("Logging on")
	const (
		isBroadcastYes = true
		isBroadcastNo  = false
	)
	var (
		invalidIP         netip.Addr
		prefixNeg         = -1
		prefix0           = 0
		prefixIPv4_1Host  = 32
		prefixIPv4_2Hosts = 31
		prefixIPv4_4Hosts = 30
		prefixIPv6_1Host  = 128
		prefixIPv6_2Hosts = 127
		prefixIPv6_4Hosts = 126
		// prefix valid for both IPv4 and IPv6
		prefix24        = 24 // “1.2.3.4/24” or “::/24”
		prefix64        = 64 // “::/64”
		noBroadcastIPv4 = netip.MustParseAddr("1.2.3.4")
		broadcastIPv4   = netip.MustParseAddr("1.2.3.7")
		// IPv4 32 - 24 = 8 least significant bits 1
		broadcast24     = netip.MustParseAddr("192.168.0.255")
		noBroadcastIPv6 = netip.MustParseAddr("::")
		broadcastIPv6   = netip.MustParseAddr("::3")
		// IPv6 128 - 64 = 64 least significant bits 1
		broadcast64 = netip.MustParseAddr("::ffff:ffff:ffff:ffff")
	)
	type args struct {
		addr          netip.Addr
		routingPrefix int
	}
	tests := []struct {
		name            string
		args            args
		wantIsBroadcast bool
	}{

		{"invalid IP", args{invalidIP, prefix24}, isBroadcastNo},
		{"invalid prefix zero", args{broadcast24, prefix0}, isBroadcastNo},
		{"invalid prefix -1", args{broadcast24, prefixNeg}, isBroadcastNo},

		{"one-address IPv4 network", args{broadcastIPv4, prefixIPv4_1Host}, isBroadcastNo},
		{"two-address IPv4 network", args{broadcastIPv4, prefixIPv4_2Hosts}, isBroadcastNo},
		{"four-address IPv4 not broadcast", args{noBroadcastIPv4, prefixIPv4_4Hosts}, isBroadcastNo},
		{"four-address IPv4 broadcast", args{broadcastIPv4, prefixIPv4_4Hosts}, isBroadcastYes},
		{"255-address IPv4 broadcast", args{broadcast24, prefix24}, isBroadcastYes},

		{"one-address IPv6 network", args{broadcastIPv6, prefixIPv6_1Host}, isBroadcastNo},
		{"two-address IPv6 network", args{broadcastIPv6, prefixIPv6_2Hosts}, isBroadcastNo},
		{"four-address IPv6 not broadcast", args{noBroadcastIPv6, prefixIPv6_4Hosts}, isBroadcastNo},
		{"four-address IPv6 broadcast", args{broadcastIPv6, prefixIPv6_4Hosts}, isBroadcastYes},
		{"64-bit IPv6 broadcast", args{broadcast64, prefix64}, isBroadcastYes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsBroadcast := IsBroadcast(tt.args.addr, tt.args.routingPrefix); gotIsBroadcast != tt.wantIsBroadcast {
				t.Errorf("IsBroadcast() = %v, want %v", gotIsBroadcast, tt.wantIsBroadcast)
			}
		})
	}
}
