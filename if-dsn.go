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
// used with [psql.DBFactory.NewDB]
//   - a data source represents a set of SQL tables,
//     possibly partitioned,
//     against which queries can be prepared and executed
//   - applications execute text queries against
//     the return-value from [psql.DBFactory.NewDB]
//   - DataSourceNamer applies to any database implementation
//   - sqliter provides database implementations for SQLite3
//   - the data source namer can map an application name and
//     partition indentifier to the data source to be used
type DataSourceNamer interface {
	// DSN returns the data source name based on a partition selector
	//	- partition: partition key that determine database name
	//	- dataSourceName: identifies the database
	//	- —
	//	- DSN is typically used internally by [psql.DBFactory.NewDB]
	//	- upon creation, the data source namer was provided with
	//		information on data source naming for a particular application program
	DSN(partition ...DBPartition) (dataSourceName DataSourceName)
	// DataSource returns a usable data source based on a data source name
	//	- dsn: a partition-keyed data source name retrieved using [DataSourceNamer.DNS]
	//	- dataSource: a database that can prepare statements
	//	- —
	//	- DataSource and dataSource are typically used internally by [psql.DBFactory.NewDB]
	//	- with [psql.DBFactory.NewDB], all statements are cached prepared statements
	//		for maximum performance.
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
// implement to provide a read-only flag
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

// RowScanner is used for SQL row iteration
type RowScanner[T any] interface {
	// Scan scans a row from a SQL result-set into type t
	//	- sqlRows: the result of a multiple-row query like SELECT
	//	- t: scanned value
	//	- err: any error during scanning
	//	- —
	//	- upon Scan invocation, [sql.Rows.Next] has returned true
	//		confirming that another row does exist
	//   - Scan invokes [sql.Rows.Scan] to convert the column values
	//		of an SQL row into specific type T
	Scan(sqlRows *sql.Rows) (t T, err error)
}
