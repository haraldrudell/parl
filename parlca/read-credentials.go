/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import "github.com/haraldrudell/parl"

const (
	// extension for binary keys and certificates “.der”
	//	- der: Distinguished Encoding Rules binary format
	//	- mim types:
	//		‘application/pkix-cert’ certificate
	//		‘application/pkcs8’ key
	//		RFC 2585 for a DER-encoded X.509 certificate
	DerExt = ".der"
	// extension for text
	//	- private enhanced mail
	//	- mime-types:
	//		‘application/x-pem-file’
	//		‘application/x-x509-user-cert’
	//		‘application/x-x509-ca-cert’
	PemExt = ".pem"
)

// ReadCredentials returns server credentials by reading pem files
//   - isFileNotFound true: err is nil
func ReadCredentials(certFile, keyFile string) (
	cert parl.Certificate,
	key parl.PrivateKey,
	isFileNotFound bool,
	err error,
) {
	if cert, _ /*privateKey*/, _ /*publicKey*/, err = ReadPemFromFile(certFile, NotFoundNotError); err != nil {
		return
	} else if cert == nil {
		isFileNotFound = true
		return
	} else if _ /*cert*/, key, _ /*publicKey*/, err = ReadPemFromFile(certFile, NotFoundNotError); err != nil {
		return
	} else if key == nil {
		isFileNotFound = true
	}

	return
}
