/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"database/sql"
)

type DB interface {
	Exec(partition DBPartition, query string, ctx context.Context,
		args ...any) (execResult ExecResult, err error)
	Query(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRows *sql.Rows, err error)
	QueryRow(partition DBPartition, query string, ctx context.Context,
		args ...any) (sqlRow *sql.Row, err error)
	QueryString(partition DBPartition, query string, ctx context.Context,
		args ...any) (value string, err error)
	QueryInt(partition DBPartition, query string, ctx context.Context,
		args ...any) (value int, err error)
	Close() (err error)
}

type ExecResult interface {
	Get() (ID int64, rows int64)
	String() (s string)
}

type DBPartition string

// DBFactory is a standardized way to obtain DB objects
type DBFactory interface {
	NewDB(
		dsnr DataSourceNamer,
		schema func(dataSource DataSource, ctx context.Context) (err error)) (db DB)
}
