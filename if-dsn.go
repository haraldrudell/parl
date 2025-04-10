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
//   - sqliter provides implementations for SQLite3
//   - the data source namer can map an application name and
//     partition indentifier to the data source to be used
type DataSourceNamer interface {
	// DSN returns the data source name based on a partition selector
	//	- upon creation, the data source namer was provided with
	//		information on data source naming for a particular application program
	DSN(partition ...DBPartition) (dataSourceName DataSourceName)
	// DataSource returns a usable data source based on a data source name
	//	- with parl, all statements are prepared statements.
	//		The function provided by a data source is to create prepared statements.
	//		Those prepared statements are later executed efficiently
	DataSource(dsn DataSourceName) (dataSource DataSource, err error)
}

// DataSource is a value referring to a set of SQL tables,
// possibly a partition
//   - a datasource is typically implemented by [sql.DB]
//     delegating to the SQL driver in use
//   - a data source name is SQL driver dependent
//   - for SQLite3, a data source name is a filename like
//     “~/.local/share/myapp/myapp-2024.db”
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

// IsRODsnr is an optional interface a data source namer can
// impleement to provide a read-only flag
type IsRoDsnr interface {
	// IsRO returns true if this data source is read-only
	//	- SQLite3 will not create database files
	//	- ORM will not write schema for new database files
	IsRO() (isRO ROtype)
}

const (
	// the datasource is read-only
	ROyes = iota
	// the data source is read-write
	ROno
)

// ROtype indicates read-only: [ROno] [ROyes]
type ROtype uint8
