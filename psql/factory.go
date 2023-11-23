/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"

	"github.com/haraldrudell/parl"
)

// DBFactory implements [parl.DBFactory] providing:
//   - cached database access
//   - cached prepared statement queries
var DBFactory = &dbFactory{}

type dbFactory struct{}

// NewDB returns a [parl.DBFactory] implementation caching
// database access and prepared statements
//   - dsnr is a database-implementation data-source namer
//   - schema is a database-implementation schema initializer
func (df *dbFactory) NewDB(
	dsnr parl.DataSourceNamer,
	schema func(dataSource parl.DataSource, ctx context.Context) (err error),
) (db parl.DB) {
	return NewDBMap(dsnr, schema)
}
