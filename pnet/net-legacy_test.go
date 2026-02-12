/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"testing"
)

func TestIsIPv4(t *testing.T) {
	const (
		isIPv4Yes = true
		isIPv4No  = false
	)
	var (
		nil_IP net.IP
	)
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name       string
		args       args
		wantIsIPv4 bool
	}{
		{"nil IP", args{nil_IP}, isIPv4No},
		{"bad IP", args{net.IP([]byte{0})}, isIPv4No},
		{"IPv4", args{net.IPv4zero}, isIPv4Yes},
		{"IPv6", args{net.IPv6zero}, isIPv4No},
		{"IPv6", args{net.IPv6zero}, isIPv4No},
		{"IPv4-mapped", args{net.IPv4zero.To16()}, isIPv4Yes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsIPv4 := IsIPv4(tt.args.ip); gotIsIPv4 != tt.wantIsIPv4 {
				t.Errorf("IsIPv4() = %v, want %v", gotIsIPv4, tt.wantIsIPv4)
			}
		})
	}
}

func TestMaskToBits(t *testing.T) {
	const (
		zeroOneBits = 0
		ipv6No      = false
		ipv6Yes     = true
		errorNo     = false
		errorYes    = true
		ones24      = 24
		ones64      = 64
	)
	var (
		maskNil            net.IPMask
		ipv4SlashEight     = net.IPMask{255, 255, 255, 0}
		ipv6SlashSixtyFour = net.IPMask{
			255, 255, 255, 255,
			255, 255, 255, 255,
			0, 0, 0, 0,
			0, 0, 0, 0,
		}
	)
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		mask    net.IPMask
		want    int
		want2   bool
		wantErr bool
	}{
		{"nil-slice", maskNil, zeroOneBits, ipv6No, errorYes},
		{"ipv4/8", ipv4SlashEight, ones24, ipv6No, errorNo},
		{"ipv6/64", ipv6SlashSixtyFour, ones64, ipv6Yes, errorNo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2, gotErr := MaskToBits(tt.mask)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MaskToBits() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MaskToBits() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("MaskToBits() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("MaskToBits() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
