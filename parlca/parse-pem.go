/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// ParsePem reads the content in a pem-format byte sequence
//   - pemData: text pem-format data byte-sequence
//   - certificate: non-nil if the first pem-block successfully parsed as a “CERTIFICATE”
//   - privateKey: non-nil if the first pem contained a pkcs8 “PRIVATE KEY”
//   - publicKey: non-nil if the first pem contained a pkix encoded “PUBLIC KEY”
//   - can do rsa, ecdsa, ed25519 keys and x.509 certificates
//   - reads the first pem-block present
//   - errors:
//   - — no pem-block found
//   - — pem parsing failed
//   - — a different pem block type was encountered
func ParsePem(pemData parl.PemBytes) (
	certificate parl.Certificate,
	privateKey parl.PrivateKey,
	publicKey parl.PublicKey,
	err error,
) {

	// decode the pem block to obtain its type
	var block, _ = /*rest*/ pem.Decode(pemData)
	if block == nil {
		err = perrors.NewPF("PEM block not found in input")
		return
	}
	switch block.Type {
	case pemPublicKeyType, pemRSAPublicKeyType:

		// “PUBLIC KEY”
		publicKey, err = parsePkix(block.Bytes)
	case pemPrivateKeyType:

		// “PRIVATE KEY”
		privateKey, err = parsePkcs8(block.Bytes)
	case pemCertificateType:

		// “CERTIFICATE”
		if _, err = x509.ParseCertificate(block.Bytes); perrors.IsPF(&err, "x509.ParseCertificate %w", err) {
		}
		certificate = &Certificate{der: block.Bytes}
	default:
		err = perrors.ErrorfPF("Unknown pem block type: %q", block.Type)
	}

	return
}
