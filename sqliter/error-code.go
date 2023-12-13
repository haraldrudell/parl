/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"errors"
)

// SQLite errors are of type sqlite.Error
//   - sqlite.Error has a code int value
//   - Some code values are not exported
//   - For calling code to not have to import the driver itself,
//     frequent error-code values are here
const (
	CodeBusy                    = 5    // sqlite3.SQLITE_BUSY
	CodeConstraintFailedNOTNULL = 1299 // SQLITE_CONSTRAINT_NOTNULL
	CodeConstraintFailedUNIQUE  = 2067 // SQLITE_CONSTRAINT_UNIQUE
	CodeDatabaseIsLocked        = 261  // locked WAL file
)

// ErrorCode is the SQLite error implementation
//   - ErrorCode is an error instance
//   - The Code method provides an int error code
type ErrorCode interface {
	error
	Code() (code int)
}

// GetErrorCode traverses an error chain looking for an SQLite error
//   - If an SQLite error is found, it is returned in sqliteError
//   - — code is the int error code
//   - If no SQLite error exists in the error chain, sqliteError is nil and code is 0
func Code(err error) (code int, sqliteError ErrorCode) {
	if errors.As(err, &sqliteError) {
		code = sqliteError.Code()
	}

	return
}
