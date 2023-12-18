/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"github.com/google/uuid"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// length 36 by storing uuid.UUID as SQLite TEXT
	duUUIDlen = len("01234567-89ab-cdef-0123-456789abcdef")
)

// UUIDToDB stores a Go 128-bit UUID as 36 characters SQLite TEXT
//   - Go reflect type: array [16]byte from package github.com/google/uuid
//   - — 16×8 is 128 bits
//   - SQLite type: 36-character string TEXT “01234567-89ab-cdef-0123-456789abcdef”
//   - — 32×4 is 128 bits
//   - — SQLite [Storage Classes]: NULL INTEGER REAL TEXT BLOB
//
// [Storage Classes]: https://sqlite.org/datatype3.html#storage_classes_and_datatypes
func UUIDToDB(ID uuid.UUID) (dbValue string) {
	// 128-bit integer stored as string
	//	- 36 characters: 32 lower-case hexadecimal characters 4-bits each: 0-f and 4 hyphens “-”
	//	- “xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx”
	//	- “01234567-89ab-cdef-0123-456789abcdef”
	return ID.String()
}

// ToUUID converts an SQLite TEXT value to Go uuid.UUID
//   - err: non-nil on failure
//   - ID: valid or zero-value on failure
//   - Go reflect type: [16]byte uuid.UUID from package github.com/google/uuid
//   - — 16×8 is 128 bits
//   - SQLite value: 36-character TEXT “01234567-89ab-cdef-0123-456789abcdef”
//   - — 32×4 is 128 bits
func ToUUID(IDString string) (ID uuid.UUID, err error) {

	// attempt to parse IDString
	if len(IDString) == duUUIDlen {
		if ID, err = uuid.Parse(IDString); err != nil {
			err = perrors.ErrorfPF("uuid.Parse: %w", err)
		}
		return // parse return: ID: valid or zero-value, err: non-nil if parse failed
	}

	err = perrors.ErrorfPF("bad length: %d exp: %d", len(IDString), duUUIDlen)

	return // bad legth return: UUID zero-value, err non-nil
}
