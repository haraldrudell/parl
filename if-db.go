/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"database/sql"
	"fmt"
)

// NoPartition value interacts with a data source that is not partitioned by year or otherwise
//   - [parl.DB] represents multiple partitioned datasources and provides caching
//     of prepared statements
//   - a data source is a database that contains tables provided by a SQL driver implementation
const NoPartition DBPartition = ""

const (
	// [DB.QueryString] [DB.QueryInt] zero rows returns zero-value
	NoRowsOK NoRowsAction = true
	// [DB.QueryString] [DB.QueryInt] zero rows returns error
	NoRowsError NoRowsAction = false
)

// [DB.QueryString] [DB.QueryInt] noRowsOK
//   - indicates whether a zero-row result returns error
//   - NoRowsOK NoRowsError
type NoRowsAction bool

// DB is a parallel database connection
//   - [parl.DB]:
//   - — applies to any database implementation and any SQL driver
//   - — provides partition-mapping between a single [parl.DB] and
//     any number of [sql.DB] comprising a partitioned data source
//   - — hides the complexity of caching and partitioning
//     through use of a single [parl.DB] value
//   - — always uses context allowing pre-emptive cancelation of
//     SQL operations
//   - [psql.NewDB] provides a [parl.DB] implementation with:
//   - — caching of prepared statements per [sql.DB] data source and
//   - — on-the-fly schema-creation for created [sql.DB] data sources
type DB interface {

	// Exec executes a query not returning any rows
	//	- ExecResult contains last inserted ID if any and rows affected
	Exec(partition DBPartition, query string, ctx context.Context,
		args ...any) (execResult ExecResult, err error)

	// Query executes a query returning zero or more rows
	Query(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRows *sql.Rows, err error)

	// QueryRow executes a query returning only its first row
	//	- zero rows returns error: sql: no rows in result set: use [DB.Query]
	QueryRow(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRow *sql.Row, err error)

	// QueryString executes a query known to return zero or one row and first column a string value
	//   - implemented by [sql.DB.QueryRowContext]
	QueryString(partition DBPartition, query string, noRowsOK NoRowsAction, ctx context.Context,
		args ...any) (value string, hasValue bool, err error)

	// QueryInt executes a query known to return zero or one row and first column an int value
	//   - implemented by [sql.DB.QueryRowContext]
	QueryInt(partition DBPartition, query string, cnoRowsOK NoRowsAction, tx context.Context,
		args ...any) (value int, hasValue bool, err error)

	// Close closes the database connection
	Close() (err error)
	fmt.Stringer
}

// ExecResult is the result from [DB.Exec], a query not returning rows
//   - in SQL, such queries instead return last inserted ID and number of rows affected
//   - by having a new function exposing any errors,
//     Get and String methods can be error-free
//   - [psql2.NewExecResult] returns implementation
type ExecResult interface {
	// - ID is last inserted ID if any
	// - rows is number of rows affected
	Get() (ID int64, rows int64)
	// “sql.Result: ID afe3… rows: 123”
	String() (s string)
}

// DBPartition is partition reference for a partitioned database
//   - partition is typically one table per year
//   - DBPartition applies to any database implementation
//   - [NoPartition] is used for an unpartitioned database
type DBPartition string

// DBFactory is a standardized way to obtain [parl.DB] objects
//   - DBFactory applies to any database implementation
type DBFactory interface {
	// NewDB returns a DB object implementation
	//	- dsnr: a data source namer creating data sources:
	//	- [DataSourceNamer.DSN] returns the data source name for
	//		a partition identifier
	//	- [DataSourceNamer.DataSource] returns a data source
	//		for a data source name.
	//		The data source provides SQL query execution and data storage
	//	- —[sqliter.OpenDataSourceNamer] creates data source namer
	//		implementation for SQLite3
	//	- schema: a function that on execution carries out on-the-fly
	//		application-specific SQL initialization for a new datasource
	//	- — executes CREATE TABLE for tables and indexes
	//	- — configures database-specific referential integrity and journaling
	NewDB(
		dsnr DataSourceNamer,
		schema func(dataSource DataSource, ctx context.Context) (err error),
	) (db DB)
}

// Pragmas is the type of returned pragmas for an SQLite3 database
//   - is [fmt.Stringer]
type Pragmas interface {
	// pragmas aee returned in a key-value map
	Map() (pragmaMap map[string]string)
	// pragmas are returned in a space-separated key-value string
	// sorted by 8-bit characters
	//	- “foreignKeys: 0 journal: memory timeout: 0”
	fmt.Stringer
}
