/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"io/fs"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/pos"
)

// DataSourceNamerRO is a SQLite3 data source that
// does not create files or directories
type DataSourceNamerRO struct {
	DataSourceNamer
}

// OpenDataSourceNamerRO returns a SQLIte3 [parl.DataSourceNamer]
// that does not create databases or write to the file-system
//   - errors can be checked to be cause by read-only:
//
// Usage:
//
//	if errors.Is(err, sqliter.ErrDsnNotExist) {
func OpenDataSourceNamerRO(appName string) (dsn parl.DataSourceNamer, err error) {

	// ensure that apppName’s directory already exists
	var appDir string
	// Path retrieves directory without panics or file-system writes
	if appDir, _, err = pos.NewAppDir(appName).Path(); err != nil {
		return // some error
	}
	var fileInfo fs.FileInfo
	var isNotExist bool
	if fileInfo, isNotExist, err = pfs.Exists2(appDir); err != nil {
		if isNotExist {
			err = MarkDsnNotExist(err)
		}
		return // isNotExist or some error
	} else if !fileInfo.IsDir() {
		err = MarkDsnNotExist(perrors.ErrorfPF("not directory: %q", appDir))
		return // not directory error
	}

	var n = DataSourceNamerRO{DataSourceNamer: *newDataSourceNamer(appName)}
	if err = n.DataSourceNamer.createRO(); err != nil {
		return
	}
	dsn = &n

	return
}

// invokeUserHomeDir captures panics in pos.UserHomeDir
// func invokeUserHomeDir() (homeDir string, err error) {
// 	parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

// 	homeDir = pos.UserHomeDir()

// 	return
// }

// DataSource does not create database files
//   - dataSourceName is the path toa file-syste database file
//   - make sure it already exists
func (n *DataSourceNamerRO) DataSource(dataSourceName parl.DataSourceName) (dataSource parl.DataSource, err error) {
	var isNotExist bool
	if _, isNotExist, err = pfs.Exists2(string(dataSourceName)); err != nil {
		if isNotExist {
			err = MarkDsnNotExist(err)
		}
		return // isNotExist or some error
	}

	//th edatabase file exists: open it
	return n.DataSourceNamer.DataSource(dataSourceName)
}
