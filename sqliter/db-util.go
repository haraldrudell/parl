/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

const (
	duFrames  = 1
	duUUIDlen = 36
)

type DBUtil struct{}

func (du *DBUtil) BoolToDB(b bool) (dbValue int) {
	if b {
		dbValue = 1
	}
	return
}

func (du *DBUtil) ToBool(dbValue int) (b bool, err error) {
	if dbValue != 0 && dbValue != 1 {
		err = perrors.Errorf("DBToBool: illegal value: %d exp 0 1", dbValue)
		return
	}
	if dbValue != 0 {
		b = true
	}
	return
}

func (du *DBUtil) UUIDToDB(ID uuid.UUID) (dbValue string) {
	return ID.String()
}

func (du *DBUtil) ToUUID(IDString string) (ID uuid.UUID, err error) {
	if len(IDString) != duUUIDlen {
		err = perrors.Errorf("ToUUID: bad length: %d exp: %d", len(IDString), duUUIDlen)
	}
	if ID, err = uuid.Parse(IDString); err != nil {
		err = perrors.Errorf("ParseUUID: %w", err)
	}
	return
}

func (du *DBUtil) TimeToDB(t time.Time) (dbValue string) {
	return ptime.Rfc3339nsz(t)
}

func (du *DBUtil) ToTime(timeString string) (t time.Time, err error) {
	if t, err = ptime.ParseRfc3339nsz(timeString); err != nil {
		err = perrors.Errorf("ToTime: %w", err)
		return
	}
	t = t.Local()
	return
}

func (du *DBUtil) Int(sqlRow *sql.Row, e error) (value int, err error) {
	if e != nil {
		err = e
		return
	}
	if err = sqlRow.Scan(&value); err != nil {
		err = perrors.Errorf("QueryRowYear.Scan: %v", err)
		return
	}

	return
}

func (du *DBUtil) ExecResult(id int64, rows int64, err error) (s string) {
	if err != nil {
		s = " err: " + err.Error()
	}
	return parl.Sprintf("id %d rows %d%s",
		id,
		rows,
		s,
	)
}
