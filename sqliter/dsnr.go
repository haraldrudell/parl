/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
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
	urwx      os.FileMode = 0700
	extension             = ".db"
	hyphen                = "-"
)

type DataSourceNamer struct {
	// dir is absolute path to a writable directory in user’s home directory
	dir     string
	appName string
}

func NewDataSourceNamer(appName string) (dns parl.DataSourceNamer) {
	d := DataSourceNamer{
		dir:     pos.AppDir(appName),
		appName: appName,
	}

	// create directory in user’s home based on app name
	// ~/.local/share/harvestlogs/harvestlogs-2022.db
	if err := os.MkdirAll(d.dir, urwx); err != nil {
		panic(perrors.Errorf("os.MkdirAll: %w %q", err, d.dir))
	}

	return &d
}

func (dn *DataSourceNamer) DSN(year ...parl.DBPartition) (dsn string) {

	// get database file name
	var filename string
	var year0 parl.DBPartition
	if len(year) > 0 {
		year0 = year[0]
	}
	if year0 != "" {
		filename = dn.appName + hyphen + string(year0) + extension
	} else {
		filename = dn.appName + extension
	}

	return filepath.Join(dn.dir, filename)
}

func (dn *DataSourceNamer) DataSource(dataSourceName string) (dataSource parl.DataSource, err error) {
	return NewDataSource(dataSourceName)
}
