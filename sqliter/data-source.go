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
	"github.com/haraldrudell/parl/psql"
	_ "modernc.org/sqlite"
)

const (
	sqLiteDriverName = "sqlite"
	sqStatement      = "stmt"
)

type DataSource struct {
	*sql.DB
	counters parl.Counters
}

// NewDB get a DB object that repreents the databases in a directory
func NewDataSource(dataSourceName string) (dataSource parl.DataSource, err error) {
	d := DataSource{
		counters: counter.CountersFactory.NewCounters(true, nil), // nil: no rate counters
	}

	if d.DB, err = sql.Open(sqLiteDriverName, dataSourceName); err != nil {
		err = perrors.Errorf("sql.Open(%s %s): %w", sqLiteDriverName, dataSourceName, err)
		return
	}

	dataSource = &d
	return
}

// PrepareContext returns a sql.Stmt that does retries on 5 SQLITE_BUSY
func (ds *DataSource) WrapStmt(stmt *sql.Stmt) (stm psql.Stmt) {
	return &Stmt{Stmt: stmt, ds: ds}
}
