/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import "github.com/haraldrudell/parl/perrors"

const (
	// SQLite INTEGER value for false: 0
	boolIntFalse int = 0
	// SQLite INTEGER value for true: 1
	boolIntTrue int = 1
)

// BoolToDB converts Go bool to SQLite INTEGER
//   - Go reflect type is bool
//   - SQLite storage class for [Boolean Datatype] is INTEGER
//   - SQLite values used are 0 or 1
//
// [Boolean Datatype]: https://sqlite.org/datatype3.html#boolean_datatype
func BoolToDB(b bool) (dbValue int) {
	if b {
		dbValue = boolIntTrue
	} else {
		dbValue = boolIntFalse
	}
	return
}

// ToBool converts SQLite INTEGER 0 or 1 to Go bool
//   - Go reflect type is bool
//   - SQLite INTEGER values are 0 or 1
func ToBool(dbValue int) (b bool, err error) {
	if dbValue == boolIntFalse {
		return // false return: b false, err nil
	} else if b = dbValue == boolIntTrue; b {
		return // true return: b true, err nil
	}

	err = perrors.ErrorfPF("illegal database value for boolean: %d exp %d %d",
		dbValue, boolIntFalse, boolIntTrue,
	)

	return // error return: b false, err non-nil
}
