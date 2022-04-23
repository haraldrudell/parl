/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type DB struct {
	dsnr   parl.DataSourceNamer
	lock   sync.Mutex
	m      map[string]parl.DataSource // behind lock
	schema func(dataSource parl.DataSource, ctx context.Context) (err error)
	ctx    context.Context
}

// NewDB get a DB object that repreents the databases in a directory
func NewDB(dsnr parl.DataSourceNamer, ctx context.Context,
	schema func(dataSource parl.DataSource, ctx context.Context) (err error)) (db parl.DB) {
	d := DB{
		dsnr:   dsnr,
		m:      map[string]parl.DataSource{},
		schema: schema,
		ctx:    ctx,
	}

	return &d
}

func (db *DB) DataSource(year ...string) (dataSource parl.DataSource) {
	return db.getDataSource(year...)
}

func (db *DB) Close() {
	var err error
	for _, dbFile := range db.getDBFiles() {
		if e := dbFile.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("Db.Close: %w", e))
		}
	}
	if err != nil {
		panic(err)
	}
}

// getDBs gets all cached sql.DB objects and empties the cache
func (db *DB) getDBFiles() (dataSources []parl.DataSource) {
	db.lock.Lock()
	defer db.lock.Unlock()

	for _, dataSource := range db.m {
		dataSources = append(dataSources, dataSource)
	}
	db.m = map[string]parl.DataSource{}

	return
}

// getDB gets or create a cached sql.DB object
func (db *DB) getDataSource(year ...string) (dataSource parl.DataSource) {
	var err error

	dataSourceName := db.dsnr.DSN(year...)

	db.lock.Lock()
	defer db.lock.Unlock()

	// look for existing database
	var ok bool
	if dataSource, ok = db.m[dataSourceName]; ok {
		return // existing database file
	}

	// initialize database file
	if dataSource, err = db.dsnr.DataSource(dataSourceName); err != nil {
		panic(err)
	}
	defer func() {
		if err == nil {
			db.m[dataSourceName] = dataSource
			return
		}
		if e := dataSource.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("sqlDB.Close: %w", e))
		}
		panic(err)
	}()

	if db.schema == nil {
		return // no schema: success!
	}

	// invoke schema
	if err = db.schema(dataSource, db.ctx); err != nil {
		err = perrors.Errorf("schema: %w", err)
		return
	}

	return // good return, err == nil
}
