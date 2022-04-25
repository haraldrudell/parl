/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	_ "modernc.org/sqlite"
)

const (
	sqLiteDriverName = "sqlite"
)

type DataSource struct {
	*sql.DB
}

// NewDB get a DB object that repreents the databases in a directory
func NewDataSource(dataSourceName string) (dataSource parl.DataSource, err error) {
	d := DataSource{}

	if d.DB, err = sql.Open(sqLiteDriverName, dataSourceName); err != nil {
		err = perrors.Errorf("sql.Open(%s %s): %w", sqLiteDriverName, dataSourceName, err)
		return
	}

	dataSource = &d
	return
}
