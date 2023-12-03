/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Protocol represents iana Assigned Internet Protocol Numbers for IPv4 and IPv6.
package iana

import (
	"github.com/haraldrudell/parl/ints"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/sets"
	"golang.org/x/exp/constraints"
)

// Protocol represents iana Assigned Internet Protocol Numbers for IPv4 and IPv6.
//   - Protocol is ordered
//   - Protocol implements fmt.Stringer
//   - Protocol has methods IsValid Description Int Uint8
//
// IANA [protocol-numbers]
//
// [protocol-numbers]: https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
type Protocol uint8

const (
	IPhopopt   Protocol = iota       // IPv6 Hop-by-Hop Option IPv6xh RFC8200
	IPicmp                           // Internet Control Message RFC792
	IPigmp                           // IGMP Internet Group Management RFC1112
	IPggp                            // GGP Gateway-to-Gateway RFC823
	IPv4                             // IPv4 IPv4 encapsulation RFC2003
	IPst                             // ST Stream RFC1190 RFC1819
	IPtcp                            // TCP Transmission Control RFC9293
	IPcbt                            // CBT CBT
	IPegp                            // EGP Exterior Gateway Protocol RFC888
	IPigp                            // IGP any private interior gateway
	IPbbn                            // BBN-RCC-MON BBN RCC Monitoring
	IPnvp                            // NVP-II Network Voice Protocol RFC741
	IPpup                            // PUP PUP
	IPargus                          // ARGUS (deprecated) ARGUS
	IPemcon                          // EMCON EMCON
	IPxnet                           // XNET Cross Net Debugger
	IPchaos                          // CHAOS Chaos
	IPudp                            // UDP User Datagram RFC768
	IPmux                            // MUX Multiplexing
	IPdcn                            // DCN-MEAS DCN Measurement Subsystems
	IPhmp                            // HMP Host Monitoring RFC869
	IPprm                            // PRM Packet Radio Measurement
	IPxns                            // XNS-IDP XEROX NS IDP
	IPtrunk1                         // TRUNK-1 Trunk-1
	IPtrunk2                         // TRUNK-2 Trunk-2
	IPleaf1                          // LEAF-1 Leaf-1
	IPleaf2                          // LEAF-2 Leaf-2
	IPrdp                            // RDP Reliable Data Protocol RFC908
	IPirtp                           // IRTP Internet Reliable Transaction RFC938
	IPisotp4                         // ISO-TP4 ISO Transport Protocol Class 4 RFC905
	IPnetblt                         // NETBLT Bulk Data Transfer Protocol RFC969
	IPmfe                            // MFE-NSP MFE Network Services Protocol
	IPmerit                          // MERIT-INP MERIT Internodal Protocol
	IPdccp                           // DCCP Datagram Congestion Control Protocol RFC4340
	IP3pc                            // 3PC Third Party Connect Protocol
	IPidpr                           // IDPR Inter-Domain Policy Routing Protocol
	IPxtp                            // XTP XTP
	IPddp                            // DDP Datagram Delivery Protocol
	IPidprcmtp                       // IDPR-CMTP IDPR Control Message Transport Proto
	IPtp                             // TP++ TP++ Transport Protocol
	IPil                             // IL IL Transport Protocol
	IPv6                             // IPv6 IPv6 encapsulation RFC2473
	IPsdrp                           // SDRP Source Demand Routing Protocol
	IPv6route                        // IPv6-Route Routing Header for IPv6 IPv6xh
	IPv6frag                         // IPv6-Frag Fragment Header for IPv6 IPv6xh
	IPidrp                           // IDRP Inter-Domain Routing Protocol
	IPrsvp                           // RSVP Reservation Protocol RFC2205 RFC3209
	IPgre                            // GRE Generic Routing Encapsulation RFC2784
	IPdsr                            // DSR Dynamic Source Routing Protocol RFC4728
	IPbna                            // BNA BNA
	IPesp                            // ESP Encap Security Payload IPv6xh RFC4303
	IPah                             // AH Authentication Header IPv6xh RFC4302
	IPinlsp                          // I-NLSP Integrated Net Layer Security TUBA
	IPswipe                          // SWIPE (deprecated) IP with Encryption
	IPnarp                           // NARP NBMA Address Resolution Protocol RFC1735
	IPmobile                         // MOBILE IP Mobility
	IPtlsp                           // TLSP Transport Layer Security Protocol using Kryptonet key management
	IPskip                           // SKIP SKIP
	IPv6icmp                         // IPv6-ICMP ICMP for IPv6 RFC8200
	IPv6nonxt                        // IPv6-NoNxt No Next Header for IPv6 RFC8200
	IPv6opts                         // IPv6-Opts Destination Options for IPv6 IPv6xh RFC8200
	IPanyhost                        // ANYHOST any host internal protocol
	IPcftp                           // CFTP CFTP Network Message
	IPanynw                          // ANYNW any local network
	IPsat                            // SAT-EXPAK SATNET and Backroom EXPAK
	IPkrypto                         // KRYPTOLAN Kryptolan
	IPrvd                            // RVD MIT Remote Virtual Disk Protocol
	IPippc                           // IPPC Internet Pluribus Packet Core
	IPanyfs                          // ANYFS any distributed file system
	IPsatmon                         // SAT-MON SATNET Monitoring
	IPvisa                           // VISA VISA Protocol
	IPipcv                           // IPCV Internet Packet Core Utility
	IPcpnx                           // CPNX Computer Protocol Network Executive
	IPcphb                           // CPHB Computer Protocol Heart Beat
	IPwsn                            // WSN Wang Span Network
	IPpvp                            // PVP Packet Video Protocol
	IPbrsat                          // BR-SAT-MON Backroom SATNET Monitoring
	IPsun                            // SUN-ND SUN ND PROTOCOL
	IPwbmon                          // WB-MON WIDEBAND Monitoring
	IPwbexpak                        // WB-EXPAK WIDEBAND EXPAK
	IPisoip                          // ISO-IP ISO Internet Protocol
	IPvmtp                           // VMTP VMTP
	IPsvmtp                          // SECURE-VMTP SECURE-VMTP
	IPvines                          // VINES VINES
	IPttp                            // TTP Transaction Transport Protocol, IPTM Internet Protocol Traffic Manager
	IPnsf                            // NSFNET-IGP NSFNET-IGP
	IPdgp                            // DGP Dissimilar Gateway Protocol
	IPtcf                            // TCF TCF
	IPeigrp                          // EIGRP EIGRP RFC7868
	IPospf                           // OSPFIGP OSPFIGP RFC1583 RFC2328 RFC5340
	IPrpc                            // Sprite-RPC Sprite RPC Protocol
	IPlarp                           // LARP Locus Address Resolution Protocol
	IPmtp                            // MTP Multicast Transport Protocol
	IPax25                           // AX.25 AX.25 Frames
	IPip                             // IPIP IP-within-IP Encapsulation Protocol
	IPmicp                           // MICP (deprecated) Mobile Internetworking Control Pro.
	IPscc                            // SCC-SP Semaphore Communications Sec. Pro.
	IPether                          // ETHERIP Ethernet-within-IP Encapsulation RFC3378
	IPencap                          // ENCAP Encapsulation Header RFC1241
	IPenc                            // ENC any private encryption scheme
	IPgmtp                           // GMTP GMTP RXB5
	IPifmp                           // IFMP Ipsilon Flow Management Protocol
	IPpnni                           // PNNI PNNI over IP
	IPpim                            // PIM Protocol Independent Multicast RFC7761
	IParis                           // ARIS ARIS
	IPscps                           // SCPS SCPS
	IPqnx                            // QNX QNX
	IPan                             // A/N Active Networks
	IPcomp                           // IPComp IP Payload Compression Protocol RFC2393
	IPsnp                            // SNP Sitara Networks Protocol
	IPcq                             // Compaq-Peer Compaq Peer Protocol
	IPipx                            // IPX-in-IP IPX in IP
	IPvrrp                           // VRRP Virtual Router Redundancy Protocol RFC5798
	IPpgm                            // PGM PGM Reliable Transport Protocol
	IPany0                           // any 0-hop protocol any 0-hop protocol
	IPl2tp                           // L2TP Layer Two Tunneling Protocol RFC3931
	IPddx                            // DDX D-II Data Exchange (DDX)
	IPiatp                           // IATP Interactive Agent Transfer Protocol
	IPstp                            // STP Schedule Transfer Protocol
	IPsrp                            // SRP SpectraLink Radio Protocol
	IPuti                            // UTI UTI
	IPsmp                            // SMP Simple Message Protocol
	IPsm                             // SM (deprecated) Simple Multicast Protocol
	IPptp                            // PTP Performance Transparency Protocol
	IPisis                           // ISIS ISIS over IPv4
	IPfire                           // FIRE FIRE
	IPcrtp                           // CRTP Combat Radio Transport Protocol
	IPcrudp                          // CRUDP Combat Radio User Datagram
	IPssc                            // SSCOPMCE SSCOPMCE
	IPlt                             // IPLT IPLT
	IPsps                            // SPS Secure Packet Shield
	IPpipe                           // PIPE Private IP Encapsulation within IP
	IPsctp                           // SCTP Stream Control Transmission Protocol
	IPfc                             // FC Fibre Channel Murali_Rajagopal RFC6172
	IPrsvpi                          // RSVP RSVP-E2E-IGNORE  RFC3175
	IPmobility                       // MOBILITY Mobility Header IPv6xh RFC6275
	IPudplite                        // UDPLite UDPLite RFC3828
	IPmpls                           // MPLS-in-IP MPLS-in-IP RFC4023
	IPmanet                          // manet MANET Protocols RFC5498
	IPhip                            // HIP Host Identity Protocol IPv6xh RFC7401
	IPshim6                          // Shim6 Shim6 Protocol IPv6xh RFC5533
	IPwesp                           // WESP Wrapped Encapsulating Security Payload RFC5840
	IProhc                           // ROHC Robust Header Compression RFC5858
	IPeth                            // Ethernet Ethernet RFC8986
	IPagg                            // AGGFRAG AGGFRAG encapsulation payload for ESP RFC-ietf-ipsecme-iptfs-19
	IP253      Protocol = 108 + iota // 253 Use for experimentation and testing IPv6xh RFC3692
	IP254                            // 254 Use for experimentation and testing IPv6xh RFC3692
	IP255                            // Reserved Reserved
)

// NewProtocol returns iana.Protocol for any integer value.
//   - values larger that 255 produce error testable with errors.Is(err, ints.ErrTooLarge)
//   - protocol may be invalid, ie. not an iana-assigned value, check with protocol.IsValid
//   - or use NewValidProtocol
func NewProtocol[T constraints.Integer](integer T) (protocol Protocol, err error) {

	// convert to uint8
	var u8 uint8
	if u8, err = ints.Unsigned[uint8](integer, perrors.PackFunc()); err != nil {
		return
	}

	// convert to iana.Protocol
	protocol = Protocol(u8)

	return
}

// NewProtocol returns iana.Protocol for any integer value.
//   - values larger that 255 produce error testable with errors.Is(err, ints.ErrTooLarge)
//   - protocol is valid
func NewValidProtocol[T constraints.Integer](integer T) (protocol Protocol, err error) {
	if protocol, err = NewProtocol(integer); err != nil {
		return
	}
	if !protocol.IsValid() {
		err = perrors.ErrorfPF("invalid protocol value: %d 0x%[1]x", protocol)
		return
	}

	return
}

// NewProtocol returns iana.Protocol for any integer value.
//   - if value is too large, panic
//   - protocol may be invalid, ie. not an iana-assigned value, check with protocol.IsValid
//   - or use NewValidProtocol
func NewProtocol1[T constraints.Integer](integer T) (protocol Protocol) {
	var err error
	if protocol, err = NewProtocol(integer); err != nil {
		panic(err)
	}

	return
}

func (pr Protocol) String() (s string) {
	return ianaSet.StringT(pr)
}

func (pr Protocol) Int() (protocolInt int) {
	return int(pr)
}

func (pr Protocol) Uint8() (protocolInt uint8) {
	return uint8(pr)
}

func (pr Protocol) IsValid() (isValid bool) {
	return ianaSet.IsValid(pr)
}

// Description returns a sentence describing protocol
func (pr Protocol) Description() (full string) {
	return ianaSet.Description(pr)
}

var ianaSet = sets.NewSet[Protocol]([]sets.SetElementFull[Protocol]{
	{ValueV: IPhopopt, Name: "HOPOPT", Full: "IPv6 Hop-by-Hop Option IPv6xh RFC8200"},
	{ValueV: IPicmp, Name: "ICMP", Full: "Internet Control Message RFC792"},
	{ValueV: IPigmp, Name: "IGMP", Full: "Internet Group Management RFC1112"},
	{ValueV: IPggp, Name: "GGP", Full: "Gateway-to-Gateway RFC823"},
	{ValueV: IPv4, Name: "IPv4", Full: "IPv4 encapsulation RFC2003"},
	{ValueV: IPst, Name: "ST", Full: "Stream RFC1190 RFC1819"},
	{ValueV: IPtcp, Name: "TCP", Full: "Transmission Control RFC9293"},
	{ValueV: IPcbt, Name: "CBT", Full: "CBT"},
	{ValueV: IPegp, Name: "EGP", Full: "Exterior Gateway Protocol RFC888"},
	{ValueV: IPigp, Name: "IGP", Full: "any private interior gateway"},
	{ValueV: IPbbn, Name: "BBN-RCC-MON", Full: "BBN RCC Monitoring"},
	{ValueV: IPnvp, Name: "NVP-II", Full: "Network Voice Protocol RFC741"},
	{ValueV: IPpup, Name: "PUP", Full: "PUP"},
	{ValueV: IPargus, Name: "ARGUS", Full: "(deprecated) ARGUS"},
	{ValueV: IPemcon, Name: "EMCON", Full: "EMCON"},
	{ValueV: IPxnet, Name: "XNET", Full: "Cross Net Debugger"},
	{ValueV: IPchaos, Name: "CHAOS", Full: "Chaos"},
	{ValueV: IPudp, Name: "UDP", Full: "User Datagram RFC768"},
	{ValueV: IPmux, Name: "MUX", Full: "Multiplexing"},
	{ValueV: IPdcn, Name: "DCN-MEAS", Full: "DCN Measurement Subsystems"},
	{ValueV: IPhmp, Name: "HMP", Full: "Host Monitoring RFC869"},
	{ValueV: IPprm, Name: "PRM", Full: "Packet Radio Measurement"},
	{ValueV: IPxns, Name: "XNS-IDP", Full: "XEROX NS IDP"},
	{ValueV: IPtrunk1, Name: "TRUNK-1", Full: "Trunk-1"},
	{ValueV: IPtrunk2, Name: "TRUNK-2", Full: "Trunk-2"},
	{ValueV: IPleaf1, Name: "LEAF-1", Full: "Leaf-1"},
	{ValueV: IPleaf2, Name: "LEAF-2", Full: "Leaf-2"},
	{ValueV: IPrdp, Name: "RDP", Full: "Reliable Data Protocol RFC908"},
	{ValueV: IPirtp, Name: "IRTP", Full: "Internet Reliable Transaction RFC938"},
	{ValueV: IPisotp4, Name: "ISO-TP4", Full: "ISO Transport Protocol Class 4 RFC905"},
	{ValueV: IPnetblt, Name: "NETBLT", Full: "Bulk Data Transfer Protocol RFC969"},
	{ValueV: IPmfe, Name: "MFE-NSP", Full: "MFE Network Services Protocol"},
	{ValueV: IPmerit, Name: "MERIT-INP", Full: "MERIT Internodal Protocol"},
	{ValueV: IPdccp, Name: "DCCP", Full: "Datagram Congestion Control Protocol RFC4340"},
	{ValueV: IP3pc, Name: "3PC", Full: "Third Party Connect Protocol"},
	{ValueV: IPidpr, Name: "IDPR", Full: "Inter-Domain Policy Routing Protocol"},
	{ValueV: IPxtp, Name: "XTP", Full: "XTP"},
	{ValueV: IPddp, Name: "DDP", Full: "Datagram Delivery Protocol"},
	{ValueV: IPidprcmtp, Name: "IDPR-CMTP", Full: "IDPR Control Message Transport Proto"},
	{ValueV: IPtp, Name: "TP++", Full: "TP++ Transport Protocol"},
	{ValueV: IPil, Name: "IL", Full: "IL Transport Protocol"},
	{ValueV: IPv6, Name: "IPv6", Full: "IPv6 encapsulation RFC2473"},
	{ValueV: IPsdrp, Name: "SDRP", Full: "Source Demand Routing Protocol"},
	{ValueV: IPv6route, Name: "IPv6-Route", Full: "Routing Header for IPv6 IPv6xh"},
	{ValueV: IPv6frag, Name: "IPv6-Frag", Full: "Fragment Header for IPv6 IPv6xh"},
	{ValueV: IPidrp, Name: "IDRP", Full: "Inter-Domain Routing Protocol"},
	{ValueV: IPrsvp, Name: "RSVP", Full: "Reservation Protocol RFC2205 RFC3209"},
	{ValueV: IPgre, Name: "GRE", Full: "Generic Routing Encapsulation RFC2784"},
	{ValueV: IPdsr, Name: "DSR", Full: "Dynamic Source Routing Protocol RFC4728"},
	{ValueV: IPbna, Name: "BNA", Full: "BNA"},
	{ValueV: IPesp, Name: "ESP", Full: "Encap Security Payload IPv6xh RFC4303"},
	{ValueV: IPah, Name: "AH", Full: "Authentication Header IPv6xh RFC4302"},
	{ValueV: IPinlsp, Name: "I-NLSP", Full: "Integrated Net Layer Security TUBA"},
	{ValueV: IPswipe, Name: "SWIPE", Full: "(deprecated) IP with Encryption"},
	{ValueV: IPnarp, Name: "NARP", Full: "NBMA Address Resolution Protocol RFC1735"},
	{ValueV: IPmobile, Name: "MOBILE", Full: "IP Mobility"},
	{ValueV: IPtlsp, Name: "TLSP", Full: "Transport Layer Security Protocol using Kryptonet key management"},
	{ValueV: IPskip, Name: "SKIP", Full: "SKIP"},
	{ValueV: IPv6icmp, Name: "IPv6-ICMP", Full: "ICMP for IPv6 RFC8200"},
	{ValueV: IPv6nonxt, Name: "IPv6-NoNxt", Full: "No Next Header for IPv6 RFC8200"},
	{ValueV: IPv6opts, Name: "IPv6-Opts", Full: "Destination Options for IPv6 IPv6xh RFC8200"},
	{ValueV: IPanyhost, Name: "ANYHOST", Full: "any host internal protocol"},
	{ValueV: IPcftp, Name: "CFTP", Full: "CFTP Network Message"},
	{ValueV: IPanynw, Name: "ANYNW", Full: "any local network"},
	{ValueV: IPsat, Name: "SAT-EXPAK", Full: "SATNET and Backroom EXPAK"},
	{ValueV: IPkrypto, Name: "KRYPTOLAN", Full: "Kryptolan"},
	{ValueV: IPrvd, Name: "RVD", Full: "MIT Remote Virtual Disk Protocol"},
	{ValueV: IPippc, Name: "IPPC", Full: "Internet Pluribus Packet Core"},
	{ValueV: IPanyfs, Name: "ANYFS", Full: "any distributed file system"},
	{ValueV: IPsatmon, Name: "SAT-MON", Full: "SATNET Monitoring"},
	{ValueV: IPvisa, Name: "VISA", Full: "VISA Protocol"},
	{ValueV: IPipcv, Name: "IPCV", Full: "Internet Packet Core Utility"},
	{ValueV: IPcpnx, Name: "CPNX", Full: "Computer Protocol Network Executive"},
	{ValueV: IPcphb, Name: "CPHB", Full: "Computer Protocol Heart Beat"},
	{ValueV: IPwsn, Name: "WSN", Full: "Wang Span Network"},
	{ValueV: IPpvp, Name: "PVP", Full: "Packet Video Protocol"},
	{ValueV: IPbrsat, Name: "BR-SAT-MON", Full: "Backroom SATNET Monitoring"},
	{ValueV: IPsun, Name: "SUN-ND", Full: "SUN ND PROTOCOL"},
	{ValueV: IPwbmon, Name: "WB-MON", Full: "WIDEBAND Monitoring"},
	{ValueV: IPwbexpak, Name: "WB-EXPAK", Full: "WIDEBAND EXPAK"},
	{ValueV: IPisoip, Name: "ISO-IP", Full: "ISO Internet Protocol"},
	{ValueV: IPvmtp, Name: "VMTP", Full: "VMTP"},
	{ValueV: IPsvmtp, Name: "SECURE-VMTP", Full: "SECURE-VMTP"},
	{ValueV: IPvines, Name: "VINES", Full: "VINES"},
	{ValueV: IPttp, Name: "TTP", Full: "Transaction Transport Protocol / Internet Protocol Traffic Manager"},
	{ValueV: IPnsf, Name: "NSFNET-IGP", Full: "NSFNET-IGP"},
	{ValueV: IPdgp, Name: "DGP", Full: "Dissimilar Gateway Protocol"},
	{ValueV: IPtcf, Name: "TCF", Full: "TCF"},
	{ValueV: IPeigrp, Name: "EIGRP", Full: "EIGRP RFC7868"},
	{ValueV: IPospf, Name: "OSPFIGP", Full: "OSPFIGP RFC1583 RFC2328 RFC5340"},
	{ValueV: IPrpc, Name: "Sprite-RPC", Full: "Sprite RPC Protocol"},
	{ValueV: IPlarp, Name: "LARP", Full: "Locus Address Resolution Protocol"},
	{ValueV: IPmtp, Name: "MTP", Full: "Multicast Transport Protocol"},
	{ValueV: IPax25, Name: "AX.25", Full: "AX.25 Frames"},
	{ValueV: IPip, Name: "IPIP", Full: "IP-within-IP Encapsulation Protocol"},
	{ValueV: IPmicp, Name: "MICP", Full: "(deprecated) Mobile Internetworking Control Pro."},
	{ValueV: IPscc, Name: "SCC-SP", Full: "Semaphore Communications Sec. Pro."},
	{ValueV: IPether, Name: "ETHERIP", Full: "Ethernet-within-IP Encapsulation RFC3378"},
	{ValueV: IPencap, Name: "ENCAP", Full: "Encapsulation Header RFC1241"},
	{ValueV: IPenc, Name: "ENC", Full: "any private encryption scheme"},
	{ValueV: IPgmtp, Name: "GMTP", Full: "GMTP RXB5"},
	{ValueV: IPifmp, Name: "IFMP", Full: "Ipsilon Flow Management Protocol"},
	{ValueV: IPpnni, Name: "PNNI", Full: "PNNI over IP"},
	{ValueV: IPpim, Name: "PIM", Full: "Protocol Independent Multicast RFC7761"},
	{ValueV: IParis, Name: "ARIS", Full: "ARIS"},
	{ValueV: IPscps, Name: "SCPS", Full: "SCPS"},
	{ValueV: IPqnx, Name: "QNX", Full: "QNX"},
	{ValueV: IPan, Name: "A/N", Full: "Active Networks"},
	{ValueV: IPcomp, Name: "IPComp", Full: "IP Payload Compression Protocol RFC2393"},
	{ValueV: IPsnp, Name: "SNP", Full: "Sitara Networks Protocol"},
	{ValueV: IPcq, Name: "Compaq-Peer", Full: "Compaq Peer Protocol"},
	{ValueV: IPipx, Name: "IPX-in-IP", Full: "IPX in IP"},
	{ValueV: IPvrrp, Name: "VRRP", Full: "Virtual Router Redundancy Protocol RFC5798"},
	{ValueV: IPpgm, Name: "PGM", Full: "PGM Reliable Transport Protocol"},
	{ValueV: IPany0, Name: "any 0-hop protocol", Full: "any 0-hop protocol"},
	{ValueV: IPl2tp, Name: "L2TP", Full: "Layer Two Tunneling Protocol RFC3931"},
	{ValueV: IPddx, Name: "DDX", Full: "D-II Data Exchange (DDX)"},
	{ValueV: IPiatp, Name: "IATP", Full: "Interactive Agent Transfer Protocol"},
	{ValueV: IPstp, Name: "STP", Full: "Schedule Transfer Protocol"},
	{ValueV: IPsrp, Name: "SRP", Full: "SpectraLink Radio Protocol"},
	{ValueV: IPuti, Name: "UTI", Full: "UTI"},
	{ValueV: IPsmp, Name: "SMP", Full: "Simple Message Protocol"},
	{ValueV: IPsm, Name: "SM", Full: "(deprecated) Simple Multicast Protocol"},
	{ValueV: IPptp, Name: "PTP", Full: "Performance Transparency Protocol"},
	{ValueV: IPisis, Name: "ISIS", Full: "ISIS over IPv4"},
	{ValueV: IPfire, Name: "FIRE", Full: "FIRE"},
	{ValueV: IPcrtp, Name: "CRTP", Full: "Combat Radio Transport Protocol"},
	{ValueV: IPcrudp, Name: "CRUDP", Full: "Combat Radio User Datagram"},
	{ValueV: IPssc, Name: "SSCOPMCE", Full: "SSCOPMCE"},
	{ValueV: IPlt, Name: "IPLT", Full: "IPLT"},
	{ValueV: IPsps, Name: "SPS", Full: "Secure Packet Shield"},
	{ValueV: IPpipe, Name: "PIPE", Full: "Private IP Encapsulation within IP"},
	{ValueV: IPsctp, Name: "SCTP", Full: "Stream Control Transmission Protocol"},
	{ValueV: IPfc, Name: "FC", Full: "Fibre Channel Murali_Rajagopal RFC6172"},
	{ValueV: IPrsvpi, Name: "RSVP", Full: "RSVP-E2E-IGNORE  RFC3175"},
	{ValueV: IPmobility, Name: "MOBILITY", Full: "Mobility Header IPv6xh RFC6275"},
	{ValueV: IPudplite, Name: "UDPLite", Full: "UDPLite RFC3828"},
	{ValueV: IPmpls, Name: "MPLS-in-IP", Full: "MPLS-in-IP RFC4023"},
	{ValueV: IPmanet, Name: "manet", Full: "MANET Protocols RFC5498"},
	{ValueV: IPhip, Name: "HIP", Full: "Host Identity Protocol IPv6xh RFC7401"},
	{ValueV: IPshim6, Name: "Shim6", Full: "Shim6 Protocol IPv6xh RFC5533"},
	{ValueV: IPwesp, Name: "WESP", Full: "Wrapped Encapsulating Security Payload RFC5840"},
	{ValueV: IProhc, Name: "ROHC", Full: "Robust Header Compression RFC5858"},
	{ValueV: IPeth, Name: "Ethernet", Full: "Ethernet RFC8986"},
	{ValueV: IPagg, Name: "AGGFRAG", Full: "AGGFRAG encapsulation payload for ESP RFC-ietf-ipsecme-iptfs-19"},
	{ValueV: IP253, Name: "253", Full: "Use for experimentation and testing IPv6xh RFC3692"},
	{ValueV: IP254, Name: "254", Full: "Use for experimentation and testing IPv6xh RFC3692"},
	//{ValueV: IP255, Name: "Reserved", Full: "Reserved"},
})
