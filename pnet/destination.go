/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"net"
	"strconv"
)

// Destination contains IP, Zone and Mask
type Destination struct {
	net.IPAddr
	net.IPMask
}

// NewDestination instantiates Destination
func NewDestination(IPAddr *net.IPAddr, IPMask *net.IPMask) (d *Destination) {
	d = &Destination{}
	if IPAddr != nil {
		d.IPAddr = *IPAddr
	}
	if IPMask != nil {
		d.IPMask = *IPMask
	}
	return
}

// Key is a string suitable as a key in a map
func (d Destination) Key() string {
	ones, _ := d.IPMask.Size() // the suffix: /24 or /32 or such of CIDR
	return d.IPAddr.String() + "/" + strconv.Itoa(ones)
}

func (d Destination) String() (s string) {
	if len(d.IPAddr.IP.To4()) == net.IPv4len {
		s = shorten(d.IPAddr.IP)
	} else {
		s = d.IPAddr.String()
	}
	ones, _ := d.IPMask.Size() // the suffix: /24 or /32 or such of CIDR
	s += "/" + strconv.Itoa(ones)
	return
}
