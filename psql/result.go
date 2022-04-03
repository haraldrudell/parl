/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"database/sql"
	"strconv"
)

// Result makes sql.Result printable
type Result struct {
	sql.Result
}

// Get obtains last id and number of affected rows
func (r *Result) Get() (id int64, count int64, sErr string, err error) {
	id, err = r.LastInsertId()
	if err == nil {
		count, err = r.RowsAffected()
		if err != nil {
			sErr = "sql.Result.RowsAffected error: " + err.Error()
		}
	} else {
		sErr = "sql.Result.LastInsertId error: " + err.Error()
	}
	return
}

func (r Result) String() (s string) {
	id, count, sErr, _ := r.Get()

	s = "sql.Result: "
	if sErr != "" {
		s += sErr
	} else {
		s += "last inserted id: " + strconv.FormatInt(id, 10) +
			"rows affected: " + strconv.FormatInt(count, 10)
	}
	return
}
