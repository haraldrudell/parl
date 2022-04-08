/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/haraldrudell/parl/perrors"
)

const (
	byteLength = 32
)

// Sha256 contains sha-256 hash
type Sha256 []byte

// Sha256Context get sha256 of a file with context
func Sha256Context(ctx context.Context, filename string) (s2 Sha256, err error) {
	var file *os.File
	if file, err = os.Open(filename); err != nil {
		return
	}
	defer func() {
		if e := file.Close(); e != nil {
			err = perrors.AppendError(err, e)
		}
	}()
	hash := sha256.New()
	if _, err = io.Copy(hash, NewContextReader(ctx, file)); err != nil {
		return
	}
	s2 = hash.Sum(nil)
	return
}

// Valid determines if hash is present
func (s2 *Sha256) Valid() bool {
	return s2 != nil && len(*s2) == byteLength
}

func (s2 Sha256) String() string {
	return hex.EncodeToString(s2)
}
