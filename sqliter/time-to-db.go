/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"database/sql"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

const (
	nsUTCLength = len("2006-01-02T15:04:05.000000000Z")
)

// TimeToDB converts ns-resolution Go time.Time in any time.Location to
// SQLite TEXT ISO8601 nano-second resolution UTC time zone
//   - SQLite TEXT: “2022-01-01T08:00:00.000000000Z”
func TimeToDB(t time.Time) (dbValue string) {
	return ptime.Rfc3339nsz(t) // “2022-01-01T08:00:00.000000000Z”
}

// ToTime parses SQLite TEXT to Go time.Time in Local location
//   - SQLite TEXT ISO8601 nano-second resolution UTC time zone
//   - SQLite TEXT: “2022-01-01T08:00:00.000000000Z”
func ToTime(timeString string) (t time.Time, err error) {

	// check length “2006-01-02T15:04:05.000000000Z”
	if len(timeString) != nsUTCLength {
		err = perrors.ErrorfPF("bad length: %d exp %d", len(timeString), nsUTCLength)
		return // bad length: t: zero value err: non-nil
	}

	// parse “2006-01-02T15:04:05.000000000Z”
	if t, err = ptime.ParseRfc3339nsz(timeString); err != nil {
		err = perrors.ErrorfPF("time.Parse: %w", err)
		return // bad parse: t: zero-value err: non-nil
	}
	t = t.Local()

	return // good return: t: time in Local, err: nil
}

// TimeToDBNullable converts ns-resolution Go time.Time in any time.Location to
// SQLite TEXT ISO8601 nano-second resolution UTC time zone
//   - for nullable database column
//   - SQLite TEXT: “2022-01-01T08:00:00.000000000Z”
func TimeToDBNullable(t time.Time) (dbValue any) {
	if t.IsZero() {
		return nil // empty string
	}
	return ptime.Rfc3339nsz(t)
}

// NullableToTime parses SQLite TEXT to Go time.Time in Local location
//   - nullable database column
//   - SQLite TEXT ISO8601 nano-second resolution UTC time zone
//   - SQLite TEXT: “2022-01-01T08:00:00.000000000Z”
func NullableToTime(nullString sql.NullString) (t time.Time, err error) {
	if !nullString.Valid {
		return // NULL: t.IsZero()
	}
	return ToTime(nullString.String)
}
