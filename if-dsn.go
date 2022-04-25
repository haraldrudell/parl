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
	DSN(partition ...DBPartition) (dataSourceName string)
	DataSource(dsn string) (dataSource DataSource, err error)
}

type DataSource interface {
	PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error)
	Close() (err error)
}

type DSNrFactory interface {
	NewDSNr(appName string) (dsnr DataSourceNamer)
}
