/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// iana provides Address Family Numbers for the Internet.
package iana

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pslices"
	"github.com/haraldrudell/parl/set"
)

// iana provides Address Family Numbers for the Internet.
// IANA [address-family-numbers]
//
// [address-family-numbers]: https://www.iana.org/assignments/address-family-numbers/address-family-numbers.xhtml
type AddressFamily uint16

const (
	AFreserved0 AddressFamily = iota         // 0 Reserved
	AFip                                     // IP (IP version 4)
	AFip6                                    // IP6 (IP version 6)
	AFnsap                                   // NSAP NSAP
	AFhdlc                                   // HDLC HDLC (8-bit multidrop)
	AFbbn                                    // BBN BBN 1822
	AF802                                    // 802 (includes all 802 media plus Ethernet
	AFe163                                   // E.163 E.163
	AFe164                                   // E.164 E.164 (SMDS, Frame Relay, ATM)
	AFf69                                    // F.69 F.69 (Telex)
	AFx121                                   // X.121 X.121 (X.25, Frame Relay)
	AFips                                    // IPX IPX
	AFappletalk                              // Appletalk Appletalk
	AFdecnet                                 // DECNET Decnet IV
	AFbv                                     // BV Banyan Vines
	AFe164n                                  // E.164N E.164 with NSAP format subaddress
	AFdns                                    // DNS DNS (Domain Name System)
	AFdn                                     // DN Distinguished Name
	AFas                                     // AS AS Number
	AFxtpip                                  // XTPIP XTP over IP version 4
	AFxtpip6                                 // XTPIP6 XTP over IP version 6
	AFxtp                                    // XTP XTP native mode XTP
	AFwwp                                    // WWP Fibre Channel World-Wide Port Name
	AFwwn                                    // WWN Fibre Channel World-Wide Node Name
	AFgwid                                   // GWID GWID
	AFafi                                    // AFI AFI for L2VPN information	[RFC4761][RFC6074]
	AFmplss                                  // MPLSS MPLS-TP Section Endpoint Identifier	[RFC7212]
	AFmplse                                  // MPLSE MPLS-TP LSP Endpoint Identifier	[RFC7212]
	AFmplsp                                  // MPLSP MPLS-TP Pseudowire Endpoint Identifier	[RFC7212]
	AFmtip                                   // MTIP MT IP: Multi-Topology IP version 4	[RFC7307]
	AFmtip6                                  // MTIP6 MT IPv6: Multi-Topology IP version 6	[RFC7307]
	AFbgp                                    // BGP BGP SFC	[RFC9015]
	AFeigrp     AddressFamily = 16352 + iota // EIGRP EIGRP Common Service Family
	AFeigrp4                                 // EIGRP4 EIGRP IPv4 Service Family
	AFeigrp6                                 // EIGRP6 EIGRP IPv6 Service Family
	AFlisp                                   // LISP LISP Canonical Address Format (LCAF)
	AFbgpls                                  // BGP-LS BGP-LS	[RFC7752]
	AFmac48                                  // MAC48 48-bit MAC	[RFC7042]	2013-05-06
	AFmac64                                  // MAC64 64-bit MAC	[RFC7042]	2013-05-06
	AFoui                                    // OUI OUI	[RFC7961]
	AFmac24                                  // MAC/24 MAC/24	[RFC7961]
	AFmac40                                  // MAC/40 MAC/40	[RFC7961]
	AFipv664                                 // IPv6/64 IPv6/64	[RFC7961]
	AFrb                                     // RBridge RBridge Port ID	[RFC7961]
	AFtrill                                  // TRILL TRILL Nickname	[RFC7455]
	AFuuid                                   // UUID Universally Unique Identifier (UUID)
	AFafir                                   // AFI Routing Policy AFI	[draft-ietf-idr-rpd-02]
	AFmplsns                                 // MPLSNS MPLS Namespaces
	AFreserved  AddressFamily = 65535        // 65535 65535	Reserved
)

func (af AddressFamily) String() (s string) {
	return addressFamilySet.StringT(af)
}

func (af AddressFamily) IsValid() (isValid bool) {
	return addressFamilySet.IsValid(af)
}

func (af AddressFamily) Description() (full string) {
	return addressFamilySet.Description(af)
}

var addressFamilySet = set.NewSet(pslices.ConvertSliceToInterface[
	set.SetElementFull[AddressFamily],
	parl.Element[AddressFamily],
]([]set.SetElementFull[AddressFamily]{
	{ValueV: AFreserved0, Name: "0", Full: "Reserved"},
	{ValueV: AFip, Name: "IP", Full: "(IP version 4)"},
	{ValueV: AFip6, Name: "IP6", Full: "(IP version 6)"},
	{ValueV: AFnsap, Name: "NSAP", Full: "NSAP"},
	{ValueV: AFhdlc, Name: "HDLC", Full: "HDLC (8-bit multidrop)"},
	{ValueV: AFbbn, Name: "BBN", Full: "BBN 1822"},
	{ValueV: AF802, Name: "802", Full: "(includes all 802 media plus Ethernet"},
	{ValueV: AFe163, Name: "E.163", Full: "E.163"},
	{ValueV: AFe164, Name: "E.164", Full: "E.164 (SMDS, Frame Relay, ATM)"},
	{ValueV: AFf69, Name: "F.69", Full: "F.69 (Telex)"},
	{ValueV: AFx121, Name: "X.121", Full: "X.121 (X.25, Frame Relay)"},
	{ValueV: AFips, Name: "IPX", Full: "IPX"},
	{ValueV: AFappletalk, Name: "Appletalk", Full: "Appletalk"},
	{ValueV: AFdecnet, Name: "DECNET", Full: "Decnet IV"},
	{ValueV: AFbv, Name: "BV", Full: "Banyan Vines"},
	{ValueV: AFe164n, Name: "E.164N", Full: "E.164 with NSAP format subaddress"},
	{ValueV: AFdns, Name: "DNS", Full: "DNS (Domain Name System)"},
	{ValueV: AFdn, Name: "DN", Full: "Distinguished Name"},
	{ValueV: AFas, Name: "AS", Full: "AS Number"},
	{ValueV: AFxtpip, Name: "XTPIP", Full: "XTP over IP version 4"},
	{ValueV: AFxtpip6, Name: "XTPIP6", Full: "XTP over IP version 6"},
	{ValueV: AFxtp, Name: "XTP", Full: "XTP native mode XTP"},
	{ValueV: AFwwp, Name: "WWP", Full: "Fibre Channel World-Wide Port Name"},
	{ValueV: AFwwn, Name: "WWN", Full: "Fibre Channel World-Wide Node Name"},
	{ValueV: AFgwid, Name: "GWID", Full: "GWID"},
	{ValueV: AFafi, Name: "AFI", Full: "AFI for L2VPN information	[RFC4761][RFC6074]"},
	{ValueV: AFmplss, Name: "MPLSS", Full: "MPLS-TP Section Endpoint Identifier	[RFC7212]"},
	{ValueV: AFmplse, Name: "MPLSE", Full: "MPLS-TP LSP Endpoint Identifier	[RFC7212]"},
	{ValueV: AFmplsp, Name: "MPLSP", Full: "MPLS-TP Pseudowire Endpoint Identifier	[RFC7212]"},
	{ValueV: AFmtip, Name: "MTIP", Full: "MT IP: Multi-Topology IP version 4	[RFC7307]"},
	{ValueV: AFmtip6, Name: "MTIP6", Full: "MT IPv6: Multi-Topology IP version 6	[RFC7307]"},
	{ValueV: AFbgp, Name: "BGP", Full: "BGP SFC	[RFC9015]"},
	{ValueV: AFeigrp, Name: "EIGRP", Full: "EIGRP Common Service Family"},
	{ValueV: AFeigrp4, Name: "EIGRP4", Full: "EIGRP IPv4 Service Family"},
	{ValueV: AFeigrp6, Name: "EIGRP6", Full: "EIGRP IPv6 Service Family"},
	{ValueV: AFlisp, Name: "LISP", Full: "LISP Canonical Address Format (LCAF)"},
	{ValueV: AFbgpls, Name: "BGP-LS", Full: "BGP-LS	[RFC7752]"},
	{ValueV: AFmac48, Name: "MAC48", Full: "48-bit MAC	[RFC7042]	2013-05-06"},
	{ValueV: AFmac64, Name: "MAC64", Full: "64-bit MAC	[RFC7042]	2013-05-06"},
	{ValueV: AFoui, Name: "OUI", Full: "OUI	[RFC7961]"},
	{ValueV: AFmac24, Name: "MAC/24", Full: "MAC/24	[RFC7961]"},
	{ValueV: AFmac40, Name: "MAC/40", Full: "MAC/40	[RFC7961]"},
	{ValueV: AFipv664, Name: "IPv6/64", Full: "IPv6/64	[RFC7961]"},
	{ValueV: AFrb, Name: "RBridge", Full: "RBridge Port ID	[RFC7961]"},
	{ValueV: AFtrill, Name: "TRILL", Full: "TRILL Nickname	[RFC7455]"},
	{ValueV: AFuuid, Name: "UUID", Full: "Universally Unique Identifier (UUID)"},
	{ValueV: AFafir, Name: "AFI", Full: "Routing Policy AFI	[draft-ietf-idr-rpd-02]"},
	{ValueV: AFmplsns, Name: "MPLSNS", Full: "MPLS Namespaces"},
	{ValueV: AFreserved, Name: "65535", Full: "65535	Reserved"},
}))
