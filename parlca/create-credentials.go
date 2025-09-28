/*
© 2023–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parlca

import (
	"crypto/x509"
	"net"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca/calib"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pnet"
)

// CreateCredentialsEd25519 returns TLS credentials and ca certificate and key
//   - algo:
//   - — [x509.Ed25519] smallest key size but as of 2024 not supported by browsers
//   - — [x509.RSA] the most commonly used algorithm for browsers
//   - — [x509.ECDSA]
//   - canonicalName: a name that uniquely identifies the CA.
//     Signed certificates will refer to this string to identify the authority.
//   - canonicalName [parlca.DefaultCAName] "": use default
//   - — default common-name is “[hostname]ca-241231”
//     ie. the host’s name, the word “ca” and a date6.
//     On macOS, name is used to uniquify, so it must be unique.
//   - ipsAndDomains: a list of IP literals and domain names for the certificate
//   - ipsAndDomains missing: certificate is for localhost: “127.0.0.1” “::1” “localhost”
//   - cert: binary server certificate. Typed byte-slice
//   - — to get x509 struct, use cert.ParseCertificate
//   - serverSigner: binary and [crypto.Signer] used to run the server.
//     Implementation depends on algorithm but is typically a typed byte-slice
//     with identifying methods
//   - caCert: binary private key and binary DER ASN.1 certificate. Typed byte-slice
//   - caKey the certificate authority private key. Typed byte-slice
//   - err: empty domain-name or unexpected error from signing-code
//   - —
//   - browsers do not support ed25519, so other algorithms with larger keys
//     should be used
//   - ca subject is hostname + "ca"
func CreateCredentials(
	algo x509.PublicKeyAlgorithm,
	canonicalName string,
	ipsAndDomains ...parl.AnyCount[pnet.Address],
) (
	cert parl.Certificate,
	serverSigner parl.PrivateKey,
	caCertDER parl.CertificateDer,
	caKey parl.PrivateKey,
	err error,
) {

	// template with ips and domains
	var template = x509.Certificate{}
	var ips parl.AnyCount[pnet.Address]
	if len(ipsAndDomains) > 0 {
		ips = ipsAndDomains[0]
	}
	var count int
	for pnetAddress := range ips.Seq {
		count++
		if netipAddr := pnetAddress.Addr(); netipAddr.IsValid() {
			var netIP net.IP = netipAddr.AsSlice()
			template.IPAddresses = append(template.IPAddresses, netIP)
		} else if domain := pnetAddress.String(); domain == "" {
			err = perrors.ErrorfPF("address#%d empty", count)
			return
		} else {
			template.DNSNames = append(template.DNSNames, domain)
		}
	}
	if count == 0 {
		template.IPAddresses = localHostIP
		template.DNSNames = localHostDomain
	}
	// template has IPs and domains

	// caCert is binary private key and binary DER ASN.1 certificate
	var caCert parl.CertificateAuthority
	if caCert, err = NewSelfSigned(canonicalName, algo); err != nil {
		return

		// expand certificate to [x509.Certificate[]
	} else if _ /*caX509*/, err = caCert.Validate(); err != nil {
		return

		// server private key
		// serverSigner is binary and [crypto.Signer] used to run the server
	} else if serverSigner, err = NewPrivateKey(algo); err != nil {
		return
	}
	caCertDER = caCert.DER()
	caKey = caCert.Private()

	// certificate use is server authentication
	calib.EnsureServer(&template)
	// public key for creating server certificate
	var serverPublic = serverSigner.Public()

	// have ca sign the certificate into binary DER ASN.1 form
	var certDER parl.CertificateDer
	// Sign only returns der
	certDER, err = caCert.Sign(&template, serverPublic)
	cert = NewCertificate(certDER)

	return
}

var (
	localHostIP     = []net.IP{pnet.IPv4loopback, net.IPv6loopback}
	localHostDomain = []string{"localhost"}
)
