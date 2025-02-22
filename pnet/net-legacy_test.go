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
