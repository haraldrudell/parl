/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql2

import (
	"database/sql"

	"github.com/haraldrudell/parl/perrors"
)

func ScanToInt(sqlRow *sql.Row, e error) (value int, err error) {
	if e != nil {
		err = e
		return
	}
	if err = sqlRow.Scan(&value); err != nil {
		err = perrors.Errorf("QueryRow.Scan: %v", err)
		return
	}

	return
}
