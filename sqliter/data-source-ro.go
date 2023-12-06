/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/counter"
	"github.com/haraldrudell/parl/perrors"
)

type DataSourceRO struct {
	DataSource
}

// NewDB get a DB object that represents the databases in a directory
//   - the driver’s methods are promoted like Query
//   - implements parl’s [DataSourceNamer.DataSource] for SQLite3
func OpenDataSource2(dataSourceName parl.DataSourceName) (dataSource parl.DataSource, err error) {

	d := DataSource{
		counters: counter.CountersFactory.NewCounters(true, nil), // nil: no rate counters
	}
	if d.DB, err = sql.Open(SQLiteDriverName, string(dataSourceName)); perrors.IsPF(&err, "sql.Open(%s %s): %w", SQLiteDriverName, dataSourceName, err) {
		return
	}
	dataSource = &d

	return
}
