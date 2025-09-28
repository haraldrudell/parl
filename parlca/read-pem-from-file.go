/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlca/calib"
)

const (
	NotFoundNotError = true
)

// ReadPemFromFile reads a file for a single entity
//   - either certificate, private key or public key or error
func ReadPemFromFile(filename string, allowNotFound ...bool) (
	certificate parl.Certificate,
	privateKey parl.PrivateKey,
	publicKey parl.PublicKey,
	err error,
) {

	var allowNotFound0 = len(allowNotFound) > 0 && allowNotFound[0]

	var pemBytes parl.PemBytes
	if pemBytes, err = calib.ReadFile(filename, allowNotFound0); err != nil {
		// file read error return
		return
	} else if allowNotFound0 && pemBytes == nil {
		return
	}
	return ParsePem(pemBytes)
}
