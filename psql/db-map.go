/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"
	"database/sql"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type DBMap struct {
	dsnr   parl.DataSourceNamer
	lock   sync.Mutex
	m      map[string]*DBCache // behind lock
	schema func(dataSource parl.DataSource, ctx context.Context) (err error)
}

func NewDBMap(dsnr parl.DataSourceNamer,
	schema func(dataSource parl.DataSource, ctx context.Context) (err error)) (dbMap *DBMap) {
	return &DBMap{
		dsnr:   dsnr,
		m:      map[string]*DBCache{},
		schema: schema,
	}
}

func (dm *DBMap) Exec(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (execResult parl.ExecResult, err error) {
	var dbCache *DBCache
	if dbCache, err = dm.getOrCreateDBCache(dm.dsnr.DSN(partition), ctx); err != nil {
		return
	}
	var stmt *sql.Stmt
	if stmt, err = dbCache.Stmt(query, ctx); err != nil {
		return
	}
	if execResult, err = NewExecResult(stmt.ExecContext(ctx, args...)); err != nil {
		err = perrors.Errorf("Exec: %w", err)
		return
	}

	return
}

func (dm *DBMap) Query(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (sqlRows *sql.Rows, err error) {
	var dbCache *DBCache
	if dbCache, err = dm.getOrCreateDBCache(dm.dsnr.DSN(partition), ctx); err != nil {
		return
	}
	var stmt *sql.Stmt
	if stmt, err = dbCache.Stmt(query, ctx); err != nil {
		return
	}
	if sqlRows, err = stmt.QueryContext(ctx, args...); err != nil {
		err = perrors.Errorf("Query: %w", err)
		return
	}

	return
}

func (dm *DBMap) QueryRow(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (sqlRow *sql.Row, err error) {
	var dbCache *DBCache
	if dbCache, err = dm.getOrCreateDBCache(dm.dsnr.DSN(partition), ctx); err != nil {
		return
	}
	var stmt *sql.Stmt
	if stmt, err = dbCache.Stmt(query, ctx); err != nil {
		return
	}
	sqlRow = stmt.QueryRowContext(ctx, args...)
	if err = sqlRow.Err(); err != nil {
		err = perrors.Errorf("QueryRow: %w", err)
		return
	}

	return
}

func (dm *DBMap) QueryString(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (value string, err error) {
	var dbCache *DBCache
	if dbCache, err = dm.getOrCreateDBCache(dm.dsnr.DSN(partition), ctx); err != nil {
		return
	}
	var stmt *sql.Stmt
	if stmt, err = dbCache.Stmt(query, ctx); err != nil {
		return
	}
	if err = stmt.QueryRowContext(ctx, args...).Scan(&value); err != nil {
		err = perrors.Errorf("QueryString.Scan: %w", err)
		return
	}

	return
}

func (dm *DBMap) QueryInt(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (value int, err error) {
	var dbCache *DBCache
	if dbCache, err = dm.getOrCreateDBCache(dm.dsnr.DSN(partition), ctx); err != nil {
		return
	}
	var stmt *sql.Stmt
	if stmt, err = dbCache.Stmt(query, ctx); err != nil {
		return
	}
	if err = stmt.QueryRowContext(ctx, args...).Scan(&value); err != nil {
		err = perrors.Errorf("QueryInt.Scan: %w", err)
		return
	}

	return
}

func (dm *DBMap) Close() (err error) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	if dm.dsnr == nil {
		return
	}
	dm.dsnr = nil

	for _, dbCache := range dm.m {
		if e := dbCache.Close(); e != nil {
			err = perrors.AppendError(err, e)
		}
	}
	dm.m = nil

	return
}

func (dm *DBMap) getOrCreateDBCache(dataSourceName string,
	ctx context.Context) (dbCache *DBCache, err error) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	if dm.dsnr == nil {
		err = perrors.New("Invocation after parl.DB close")
		return
	}

	if dbCache = dm.m[dataSourceName]; dbCache != nil {
		return
	}

	// create dataSource for new dbCache instance
	var dataSource parl.DataSource
	if dataSource, err = dm.dsnr.DataSource(dataSourceName); err != nil {
		return
	}
	defer func() {
		if err == nil {
			dm.m[dataSourceName] = dbCache
		} else if e := dataSource.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("dataSource.Close: %w", err))
		}
	}()

	// initialize schema
	if err = dm.schema(dataSource, ctx); err != nil {
		return
	}
	dbCache = NewDBCache(dataSource)

	return
}
