/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"database/sql"
)

// DB is a parallel database connection
//   - DB applies to any database implementation
type DB interface {
	// Exec executes a query not returning any rows
	Exec(partition DBPartition, query string, ctx context.Context,
		args ...any) (execResult ExecResult, err error)
	// Query executes a query returning zero or more rows
	Query(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRows *sql.Rows, err error)
	// Query executes a query known to return exactly one row
	QueryRow(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRow *sql.Row, err error)
	// Query executes a query known to return exactly one row and returns its string value
	QueryString(partition DBPartition, query string, ctx context.Context,
		args ...any) (value string, err error)
	// Query executes a query known to return exactly one row and returns its int value
	QueryInt(partition DBPartition, query string, ctx context.Context,
		args ...any) (value int, err error)
	// Close closes the database connection
	Close() (err error)
}

// ExecResult is the result from [DB.Exec], a query not returning rows
//   - ID, rows is how many rows were affected
type ExecResult interface {
	Get() (ID int64, rows int64)
	String() (s string)
}

// DBPartition is partition reference for a partitioned database
//   - partition is typically one table per year
//   - DBPartition applies to any database implementation
type DBPartition string

// DBFactory is a standardized way to obtain DB objects
//   - DBFactory applies to any database implementation
type DBFactory interface {
	// schema executes application-specific SQL initialization for a new datasource
	//	- executes CREATE of tables and indexes
	//	- configures database-specific referential integrity and journaling
	NewDB(
		dsnr DataSourceNamer,
		schema func(dataSource DataSource, ctx context.Context) (err error)) (db DB)
}
