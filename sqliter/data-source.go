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
	//	- default [In-Memory Databases] name is “:memory:”
	//	- if two threads or a query while another query is read,
	//		a second database connection is opened which uses a different database
	//	- from SQLite3 [version] 3.7.13+, multiple connections may
	//		share in-memory database
	//	- use URI filename: “file::memory:?cache=shared”
	//	- Go package modernc.org/sqlite 1.30.1 240606
	//		is pure Go code-compatible with SQLite 3.46.0
	//	- the in-memory filename provided here does not support
	//		partitioning.
	//		It is used for testing
	//	- because Go may open parallel database connections
	//		at any time, use of legacy filename “:memory:”
	//		produces unpredictable results
	//
	// [In-Memory Databases]: https://sqlite.org/inmemorydb.html
	// [version]: https://stackoverflow.com/questions/36447766/multiple-sqlite-connections-to-a-database-in-memory#comment60508526_36447766
	SQLiteMemoryDataSourceName = "file::memory:?cache=shared"
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
//   - dataSourceName: a filename specifying a SQLite3 database file
//   - — for an in-memory database, SQLiteMemoryDataSourceName or ":memory:" is used
//   - dataSource: wraps a [sql.DB] value
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
