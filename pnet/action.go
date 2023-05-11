/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package pnet

import "github.com/haraldrudell/parl/sets"

const (
	AddRoute        Action = iota + 1 // AddRoute describes a new route added to the routing table
	DeleteRoute                       // DeleteRoute describes a deleted route
	RouteReport                       // RouteReport describes an existing route
	Partition                         // Partition describes an existing route that may be overridden
	AddHost                           // AddHost is a host added to the routing table
	DeleteHost                        // DeleteHost is a deleted host
	HostReport                        // HostReport is an existing host
	IfAddAddr                         // IfAddAddr Add address to interface
	IfDeleteAddr                      // IfDeleteAddr Delete address from interface
	IfStatus                          // network-interface upDownStatus or down string
	AddMulticast                      // AddMulticast annouces a multicast address
	DeleteMulticast                   // DeleteMulticast annouces a multicast address disappearing
)

// Action describes the message action
type Action uint8

func (a Action) Description() (s string) {
	return actionSet.Description(a)
}

func (a Action) String() (s string) {
	return actionSet.StringT(a)
}

var actionSet = sets.NewSet(sets.NewElements[Action](
	[]sets.SetElementFull[Action]{
		{ValueV: AddHost, Name: "addHost", Full: "new host"},
		{ValueV: DeleteHost, Name: "delHost", Full: "delete host"},
		{ValueV: HostReport, Name: "host", Full: "host"},
		{ValueV: IfAddAddr, Name: "ifAddr", Full: "address"},
		{ValueV: IfDeleteAddr, Name: "ifDelAddr", Full: "delete address"},
		{ValueV: IfStatus, Name: "ifStatus", Full: "if status"},
		{ValueV: AddMulticast, Name: "ifAddM", Full: "add multicast"},
		{ValueV: DeleteMulticast, Name: "ifDelM", Full: "delete multicast"},
		{ValueV: AddRoute, Name: "addRoute", Full: "add route"},
		{ValueV: DeleteRoute, Name: "delRoute", Full: "delete route"},
		{ValueV: Partition, Name: "demoted", Full: "demoted"},
		{ValueV: RouteReport, Name: "report", Full: "report"}, // a route does not have a prefix other than 0/0 or 192.168.1.21/32
	}))
