/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/counter"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/psql/psql2"
	_ "modernc.org/sqlite"
)

const (
	// SQLite3 special filename for in-memory databases
	SQLiteMemoryDataSourceName = ":memory:"
	// name of the SQLite3 database driver
	//	- “modernc.org/sqlite”
	SQLiteDriverName = "sqlite"
)

const (
	sqStatement = "stmt"
)

// DataSource represents a SQL database that can prepare generic SQL queries
//   - implements [parl.DataSource] for SQLite3
type DataSource struct {
	// DB represents a generic SQL database that can:
	//	- offer connections
	//	- execute generuic SQL queries
	*sql.DB
	counters parl.Counters
}

// OpenDataSource creates a database-file in the file-system and
// returns its database implementation
func OpenDataSource(dataSourceName parl.DataSourceName) (dataSource parl.DataSource, err error) {

	d := DataSource{
		counters: counter.CountersFactory.NewCounters(true, nil), // nil: no rate counters
	}
	if d.DB, err = sql.Open(SQLiteDriverName, string(dataSourceName)); perrors.IsPF(&err, "sql.Open(%s %s): %w", SQLiteDriverName, dataSourceName, err) {
		return
	}
	dataSource = &d

	return
}

// PrepareContext returns a sql.Stmt that does retries on 5 SQLITE_BUSY
//   - this is used by [parl.psql]
func (ds *DataSource) WrapStmt(stmt *sql.Stmt) (stm psql2.Stmt) {
	return &Stmt{Stmt: stmt, ds: ds}
}
