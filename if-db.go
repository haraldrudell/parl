/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"database/sql"
)

// NoPartition interacts with a data source that is not partitioned by year or otherwise
const NoPartition DBPartition = ""

// DB is a parallel database connection
//   - DB applies to any database implementation
//   - psql provides implementation with caching of:
//   - — DB objects and
//   - — prepared statements
//   - DB is obtained via new function like [DBFactory.NewDB].
//     Such returned DB can use:
//   - — its data-source namer to handle partitioning
//   - — delegation to its underlying possibly partitioned DB implementation
//   - — caching of DB implementation-objects and prepared statements
//   - — its schema function to bootstrap and migrate databases
type DB interface {
	// Exec executes a query not returning any rows
	//	- ExecResult contains last inserted ID if any and rows affected
	Exec(partition DBPartition, query string, ctx context.Context,
		args ...any) (execResult ExecResult, err error)
	// Query executes a query returning zero or more rows
	Query(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRows *sql.Rows, err error)
	// Query executes a query known to return exactly one row
	//	- zero rows returns error: sql: no rows in result set
	QueryRow(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRow *sql.Row, err error)
	// Query executes a query known to return exactly one row and first column a string value
	QueryString(partition DBPartition, query string, ctx context.Context,
		args ...any) (value string, err error)
	// Query executes a query known to return exactly one row and first column an int value
	QueryInt(partition DBPartition, query string, ctx context.Context,
		args ...any) (value int, err error)
	// Close closes the database connection
	Close() (err error)
}

// ExecResult is the result from [DB.Exec], a query not returning rows
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
type DBPartition string

// DBFactory is a standardized way to obtain DB objects
//   - DBFactory applies to any database implementation
type DBFactory interface {
	// NewDB returns a DB object implementation.
	// The schema function executes application-specific SQL initialization for
	// a new datasource
	//	- executes CREATE of tables and indexes
	//	- configures database-specific referential integrity and journaling
	NewDB(
		dsnr DataSourceNamer,
		schema func(dataSource DataSource, ctx context.Context) (err error),
	) (db DB)
}
