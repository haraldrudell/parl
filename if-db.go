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
	Exec(query string, args ...any) (id int64, rows int64, err error)
	ExecContext(ctx context.Context, query string, args ...any) (id int64, rows int64, err error)
	ExecYear(year string, query string, args ...any) (id int64, rows int64, err error)
	ExecYearContext(year string, ctx context.Context, query string, args ...any) (id int64, rows int64, err error)
	Query(
		cb func(sqlRows *sql.Rows) (err error),
		query string, args ...any) (err error)
	QueryContext(ctx context.Context,
		cb func(sqlRows *sql.Rows) (err error),
		query string, args ...any) (err error)
	QueryYear(year string,
		cb func(sqlRows *sql.Rows) (err error),
		query string, args ...any) (err error)
	QueryYearContext(
		year string, ctx context.Context,
		cb func(sqlRows *sql.Rows) (err error),
		query string, args ...any) (err error)
	QueryRow(query string, args ...any) (sqlRow *sql.Row, err error)
	QueryRowContext(ctx context.Context, query string, args ...any) (sqlRow *sql.Row, err error)
	QueryRowYear(year string, query string, args ...any) (sqlRow *sql.Row, err error)
	QueryRowYearContext(year string, ctx context.Context, query string, args ...any) (sqlRow *sql.Row, err error)
	Close()
}

type DBFactory interface {
	NewDB(dsnr DataSourceNamer, ctx context.Context,
		schema func(dataSource DataSource, ctx context.Context) (err error)) (db DB)
}
