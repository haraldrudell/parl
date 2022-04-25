/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

const (
	CodeBusy = 5
)

const (
	duFrames         = 1
	duUUIDlen        = 36
	boolIntFalse int = 0
	boolIntTrue  int = 1
)

// ErrorCode is the SQLite error implementation.
// ErrorCode is an error instance.
// The Code method provides an int error code.
type ErrorCode interface {
	error
	Code() (code int)
}

// BoolToDB takes a Go bool and converts it to its SQLite representation,
// and int value 0 or 1.
func BoolToDB(b bool) (dbValue int) {
	if b {
		dbValue = 1
	}
	return
}

// ToBool takes a SQLite representation of boolen, which is int 0 or 1,
// and converts it to Go bool
func ToBool(dbValue int) (b bool, err error) {
	if dbValue != boolIntFalse && dbValue != boolIntTrue {
		err = perrors.Errorf("DBToBool: illegal value: %d exp 0 1", dbValue)
		return
	}
	if dbValue != 0 {
		b = true
	}
	return
}

// UUIDToDB task uuid value, a [16]byte from package github.com/google/uuid
// and convert it to the SQLite representation which is a 36-character string.
func UUIDToDB(ID uuid.UUID) (dbValue string) {
	return ID.String()
}

// ToUUID takes the SQLite representation of uuid, 36-char string,
// and parses it to a 16-byte uuid value for package github.com/google/uuid
func ToUUID(IDString string) (ID uuid.UUID, err error) {
	if len(IDString) != duUUIDlen {
		err = perrors.Errorf("ToUUID: bad length: %d exp: %d", len(IDString), duUUIDlen)
	}
	if ID, err = uuid.Parse(IDString); err != nil {
		err = perrors.Errorf("ParseUUID: %w", err)
	}
	return
}

// TimeToDB takes a ns-resolution typically local time zone Go time value
// and converts it to the SQLite representation which is a utc iso8601 ns string
func TimeToDB(t time.Time) (dbValue string) {
	return ptime.Rfc3339nsz(t)
}

// ToTime parses a SQLite time value, a ns iso8601 utc string,
// and parses it to time.Time in local time zone
func ToTime(timeString string) (t time.Time, err error) {
	if t, err = ptime.ParseRfc3339nsz(timeString); err != nil {
		err = perrors.Errorf("ToTime: %w", err)
		return
	}
	t = t.Local()
	return
}

// GetErrorCode traverses an error chain looking for an SQLite error.
// If an SQLite error is found, it is returned in ec.
// code is the int error code.
// If no SQLite error exist, ec is nil and code is 0.
func Code(err error) (code int, ec ErrorCode) {
	if errors.As(err, &ec) {
		code = ec.Code()
	}
	return
}
