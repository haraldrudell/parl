/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"

	"github.com/haraldrudell/parl"
)

var DBFactory = &dbFactory{}

type dbFactory struct{}

func (df *dbFactory) NewDB(dsnr parl.DataSourceNamer,
	schema func(dataSource parl.DataSource, ctx context.Context) (err error)) (db parl.DB) {
	return NewDBMap(dsnr, schema)
}
