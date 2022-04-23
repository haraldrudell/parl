/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"database/sql"
)

type DataSourceNamer interface {
	DSN(year ...string) (dataSourceName string)
	DataSource(dsn string) (dataSource DataSource, err error)
}

type DataSource interface {
	ExecContext(ctx context.Context, query string, args ...any) (id int64, rows int64, err error)
	QueryContext(ctx context.Context,
		cb func(sqlRows *sql.Rows) (err error),
		query string, args ...any) (err error)
	QueryRowContext(ctx context.Context, query string, args ...any) (sqlRow *sql.Row, err error)
	Close() (err error)
}
