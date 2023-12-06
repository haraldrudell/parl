/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import "database/sql"

// Rows is an interface implemented by sql.Rows.Next/Scan/Close
//   - used for testing
type Rows interface {
	// Next prepares the next result row for reading with the Scan method. It
	// returns true on success, or false if there is no next result row or an error
	// happened while preparing it. Err should be consulted to distinguish between
	// the two cases.
	Next() (hasRow bool)
	// Scan copies the columns in the current row into the values pointed
	// at by dest. The number of values in dest must be the same as the
	// number of columns in Rows.
	Scan(dest ...any) (err error)
	// Close closes the Rows, preventing further enumeration. If Next is called
	// and returns false and there are no further result sets,
	// the Rows are closed automatically and it will suffice to check the
	// result of Err. Close is idempotent and does not affect the result of Err.
	Close() (err error)
}

var _ Rows = &sql.Rows{}
var _ = (&sql.Rows{}).Next
var _ = (&sql.Rows{}).Scan
var _ = (&sql.Rows{}).Close
