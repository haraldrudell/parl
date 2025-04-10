/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pfs"
)

// DataSourceFile is a SQLite3 data source that
// uses a single file
type DataSourceFile struct {
	// databaseFilename is path to SQLite3 database file: “/usr/data.db”
	databaseFilename string
	// isRO is true if this is a read-only database
	isRO parl.ROtype
}

// DataSourceFile is [parl.DataSourceNamer]
var _ parl.DataSourceNamer = &DataSourceFile{}

// DataSourceFile is [parl.IsRoDsnr]
var _ parl.IsRoDsnr = &DataSourceFile{}

// - the dsn is typically used to obtain a caching database object from [psql.DBFactory.NewDB]
//   - databaseFilename: path to SQLite3 database file: “/usr/data.db”
//   - isRO: true if this is a read-only database
//   - —
//   - methods [DataSourceNamer.DSN] and [DataSourceNamer.DataSource] are used internally
//     by [parl.DB] to efficiently return possibly partitioned caching database instances
func OpenDataSourceFile(databaseFilename string, isRO ...parl.ROtype) (dsn parl.DataSourceNamer, err error) {

	// get read-only status, default read-only
	var nowRO parl.ROtype
	if len(isRO) > 0 {
		nowRO = isRO[0]
	}

	if nowRO == parl.ROyes {
		if _ /*fileInfo*/, _ /*isNotExist*/, err = pfs.Exists2(databaseFilename); err != nil {
			return
		}
	}
	dsn = &DataSourceFile{
		databaseFilename: databaseFilename,
		isRO:             nowRO,
	}
	return
}

// DSN returns the data source name based on a partition selector
//   - dsn: DSN always returns the same filename
//   - — dsn is to be used with [DataSourceFile.DataSource] for opening a database
//   - partition: unused optional parameter
func (d *DataSourceFile) DSN(partition ...parl.DBPartition) (dsn parl.DataSourceName) {
	return parl.DataSourceName(d.databaseFilename)
}

// DataSource returns a usable SQL database
//   - dsn: a value returned by [DataSourceFile.DSN]
//   - dataSource: an opened wrapper around [sql.DB]
//   - —
//   - read-only status is [DataSourceFile.IsRO]
func (d *DataSourceFile) DataSource(dsn parl.DataSourceName) (dataSource parl.DataSource, err error) {
	return OpenDataSource(dsn)
}

// IsRO returns true if this data source is read-only
//   - SQLite3 will not create database files
//   - ORM will not write schema for uninitialized database files
func (d *DataSourceFile) IsRO() (isRO parl.ROtype) { return d.isRO }
