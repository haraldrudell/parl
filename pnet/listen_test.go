package pnet_test

import (
	"net"
	"net/netip"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

func TestListen(t *testing.T) {
	// network value does not matter
	// t.Error("Logging on")
	const (
		noAdddress = ""
	)
	var (
		invalidAddrPort netip.AddrPort
	)
	type errorValue uint8
	const (
		errorYes errorValue = iota
		errorNo
	)
	type anyValue uint8
	const (
		anyYes anyValue = iota
		anyNo
	)
	type highValue uint8
	const (
		highYes highValue = iota
		highNo
	)
	tests := []struct {
		// description of this test case
		name string
		// Named input parameters for target function.
		socketAddress pnet.SocketAddress
		wantAny       anyValue
		wantAddress   string
		wantHigh      highValue
		wantErr       errorValue
	}{
		// test ‘empty-string domain’: addrPort: [::]:64801 addr: tcp [::]:64801 isAny: true isHigh true
		{"empty-string domain", pnet.NewSocketAddress(pnet.NetworkTCP, ""), anyYes, noAdddress, highYes, errorNo},
		// test ‘bad domain’: err: ‘pnet.Listen net.Listen tcp %: “listen tcp: address %: missing port in address” at pnet.Listen()-listen.go:57’
		{"bad domain", pnet.NewSocketAddress(pnet.NetworkTCP, "%"), anyYes, noAdddress, highYes, errorYes},
		// test ‘invalid addr’: err: ‘pnet.Listen net.Listen tcp invalid AddrPort: “listen tcp: address invalid AddrPort: missing port in address” at pnet.Listen()-listen.go:57’
		{"invalid addr", pnet.NewSocketAddressLiteral(pnet.NetworkTCP, invalidAddrPort), anyYes, noAdddress, highYes, errorYes},
	}
	for _, tt := range tests {
		var gotErr error
		var a net.Addr
		var addrPort netip.AddrPort
		var isAny bool
		var isHigh bool
		t.Run(tt.name, func(t *testing.T) {
			var netListener net.Listener
			netListener, gotErr = pnet.Listen(tt.socketAddress)
			if gotErr != nil {
				if tt.wantErr == errorNo {
					t.Errorf("Listen() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr == errorYes {
				t.Fatal("Listen() succeeded unexpectedly")
			}
			// is not error and did not want error

			a = netListener.Addr()
			parl.D("test ‘%s’: netListener: %t Addr: %t",
				tt.name, netListener != nil, a != nil,
			)

			// the netAddr is an address literal
			if e := netListener.Close(); e != nil {
				t.Fatal(e)
			}
			addrPort = netip.MustParseAddrPort(a.String())
			isAny = addrPort.Addr().IsUnspecified()
			isHigh = addrPort.Port() >= 32768
			if tt.wantAny == anyYes && !isAny ||
				tt.wantAny == anyNo ||
				tt.wantHigh == highYes && !isHigh ||
				tt.wantHigh == highNo {
				t.Errorf("Listen() = %v, want %v", netListener, 5)
			}
		})

		// if returned error do next test
		if gotErr != nil {
			t.Logf("test ‘%s’: err: ‘%s’", tt.name, perrors.Short(gotErr))
			continue
		}

		var network string
		if a != nil {
			network = a.Network()
		} else {
			network = "NIL"
		}
		t.Logf("test ‘%s’: addrPort: %s addr: %s %s isAny: %t isHigh %t",
			tt.name,
			addrPort, network, a, isAny, isHigh,
		)
	}
}
