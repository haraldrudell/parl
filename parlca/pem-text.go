/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlca

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/haraldrudell/parl"
)

// PemText returns lead-in text for pem format block
// - data is 1 or 2 copies of binary der data
// - — 1: only sh256 fingerprint
// - — 2: sha256 and sha1 fingerprints
func PemText(data ...[]byte) (pemText string) {
	pemText = parl.Sprintf(copyright, time.Now().Format(parl.Rfc3339s))

	// calculate sha256 fingerprint
	if len(data) > 0 {
		// [32]byte
		var hashBytes = sha256.Sum256(data[0])
		pemText += fingerPrint + hex.EncodeToString(hashBytes[:4]) + pemNewline
		if len(data) > 1 {
			var hashBytes1 = sha1.Sum(data[1])
			pemText += fingerPrint1 + hex.EncodeToString(hashBytes1[:4]) + pemNewline
		}
	}

	return
}

const (
	copyright = "Generated on %s by github.com/haraldrudell/parl\n" +
		"parl: (c) 2018-present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)\n"
	fingerPrint  = "sha256 fingerprint: "
	fingerPrint1 = "sha1 fingerprint: "
	pemNewline   = "\n"
)
