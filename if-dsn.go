/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"database/sql"
)

// DataSourceNamer provides data source names for SQL
// based on possibly appplication name and partition year
//   - a data source represents a set of SQL tables,
//     possibly partitioned,
//     against which queries can be prepared and executed
//   - DataSourceNamer applies to any database implementation
type DataSourceNamer interface {
	// DSN returns the data source name based on a partiion selector
	DSN(partition ...DBPartition) (dataSourceName DataSourceName)
	// DataSource returns a usable data source based on a data source name
	DataSource(dsn DataSourceName) (dataSource DataSource, err error)
}

// DataSource is a value referring to a set of SQL tables,
// possibly a partition
type DataSourceName string

// DataSource represents a set of SQL tables,
// possibly partitioned,
// against which queries can be prepared and executed
//   - DataSource applies to any database implementation
type DataSource interface {
	// PrepareContext prepares a statement that is a query for reading or writing data
	//	- a prepared statement can be executed multiple times
	PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error)
	// Close closes the data-source
	Close() (err error)
}

// DSNrFactory describes the signature for a data source namer new function
//   - the data source namer provides data sources for query operations
//   - DSNrFactory applies to any database implementation
type DSNrFactory interface {
	// NewDSNr returns an object that can
	//	- provide data source names from partition selectors and
	//	- provide data sources from a data source name
	DataSourceNamer(appName string) (dsnr DataSourceNamer, err error)
}
