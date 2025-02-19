/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/sets"
)

const (
	// AddRoute describes a new route added to the routing table
	AddRoute Action = iota + 1
	// DeleteRoute describes a deleted route
	DeleteRoute
	// RouteReport describes an existing route
	RouteReport
	// Partition describes an existing route that may be overridden
	Partition
	// AddHost is a host added to the routing table
	AddHost
	// DeleteHost is a deleted host
	DeleteHost
	// HostReport is an existing host
	HostReport
	// IfAddAddr Add address to interface
	IfAddAddr
	// IfDeleteAddr Delete address from interface
	IfDeleteAddr
	// network-interface upDownStatus or down string
	IfStatus
	// AddMulticast annouces a multicast address
	AddMulticast
	// DeleteMulticast annouces a multicast address disappearing
	DeleteMulticast
)

// Action describes the message action
//   - [AddRoute] [DeleteRoute] [RouteReport] [Partition]
//     [AddHost] [DeleteHost] [HostReport]
//     [IfAddAddr] [IfDeleteAddr] [IfStatus] [AddMulticast] [DeleteMulticast]
type Action uint8

// Description is short sentence describing the action
func (a Action) Description() (s string) { return actionSet.Description(a) }

// IsValid returns no error if the action is initialized and valid
func (a Action) IsValid() (err error) {
	if actionSet.IsValid(a) {
		return // valid action return
	}
	err = perrors.ErrorfPF("invalid action: %d", a)

	return
}

func (a Action) String() (s string) { return actionSet.StringT(a) }

// actionSet is [stes.Set] for [Action]
var actionSet = sets.NewSet[Action]([]sets.SetElementFull[Action]{
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
})
