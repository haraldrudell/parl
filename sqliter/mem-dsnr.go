/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import "github.com/haraldrudell/parl"

// memDataSourceNamer is a data source namer returning a SQLite3 in-memory database
type memDsnr struct{ DataSourceNamer }

// MemDsnr is a data source namer always returning an in-memory data source
var MemDsnr = &memDsnr{}

// DSN always returns the same in-memory data source
func (n *memDsnr) DSN(...parl.DBPartition) (dsn parl.DataSourceName) {
	return parl.DataSourceName(SQLiteMemoryDataSourceName)
}
