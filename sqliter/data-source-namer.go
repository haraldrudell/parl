/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"os"
	"path/filepath"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pos"
)

const (
	// created subdirectories are readableonly by the owning user
	urwx os.FileMode = 0700
	// extension is used for SQLite3 database-files
	extension = ".db"
	// hyphen is used for SQLite3 partitioned database-file naming
	hyphen = "-"
)

// DataSourceNamer provides partitioned SQLite3 database files
// within the user’s home directory
type DataSourceNamer struct {
	// dir is absolute path to a writable directory in user’s home directory
	dir string
	// appName is like “myapp”
	//	- becomes part of directory and file names
	//	- “~/.local/share/appName/appName-2022.db”
	appName string
}

// OpenDataSourceNamer is a [parl.DSNrFactory] function that returns
// a SQLite3 data-source names that returns:
//   - SQLite3 database filenames based on a partition key
//   - SQLite3 implemented data sources providing generic SQL query
//     execution
func OpenDataSourceNamer(appName string) (dsnr parl.DataSourceNamer, err error) {
	var d = newDataSourceNamer(appName)
	if err = d.create(); err != nil {
		return
	}
	dsnr = d

	return
}

func newDataSourceNamer(appName string) (namer *DataSourceNamer) {
	return &DataSourceNamer{
		appName: appName,
	}
}

// DSN returns a data source name, ie. a filename from
// application name and a partition key like year
//   - effectyively an absolute path of a writable SQLite3 “.db” database file
//   - implements parl’s [DatasourceNamer.DSN]
func (n *DataSourceNamer) DSN(year ...parl.DBPartition) (dsnr parl.DataSourceName) {
	var year0 parl.DBPartition
	if len(year) > 0 {
		year0 = year[0]
	}

	// get database file name with or without partitioning
	var filename string
	if year0 != "" {
		filename = n.appName + hyphen + string(year0) + extension
	} else {
		filename = n.appName + extension
	}
	dsnr = parl.DataSourceName(filepath.Join(n.dir, filename))

	return
}

// DataSource returns a data-source that can execute generic SQL queries
// based on a data-source name
//   - implements parl’s [DatasourceNamer.DataSource]
func (n *DataSourceNamer) DataSource(dataSourceName parl.DataSourceName) (dataSource parl.DataSource, err error) {
	return OpenDataSource(dataSourceName)
}

// create creates all necessary directories
func (n *DataSourceNamer) create() (err error) {

	if err = n.createRO(); err != nil {
		return
	}

	// create directory in user’s home based on app name
	// ~/.local/share/harvestlogs/harvestlogs-2022.db
	if err = os.MkdirAll(n.dir, urwx); perrors.IsPF(&err, "os.MkdirAll: %w %q", err, n.dir) {
		return
	}

	return
}

func (n *DataSourceNamer) createRO() (err error) {
	// determine directory
	n.dir, _, err = pos.NewAppDir(n.appName).Path()

	return
}
