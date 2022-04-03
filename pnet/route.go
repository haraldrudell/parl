/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

// Route describes a routing table route with destination and next hop
type Route struct {
	Destination
	NextHop
}

// NewRoute instantiates Route
func NewRoute(d *Destination, nextHop *NextHop) *Route {
	r := Route{}
	if d != nil {
		r.Destination = *d
	}
	if nextHop != nil {
		r.NextHop = *nextHop
	}
	return &r
}
