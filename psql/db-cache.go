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

type DBCache struct {
	ds   parl.DataSource
	lock sync.Mutex
	m    map[string]*sql.Stmt // behind lock
}

func NewDBCache(dataSource parl.DataSource) (dc *DBCache) {
	return &DBCache{
		ds: dataSource,
		m:  map[string]*sql.Stmt{},
	}
}

func (dc *DBCache) Stmt(query string, ctx context.Context) (stmt *sql.Stmt, err error) {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	if dc.ds == nil {
		err = perrors.New("Stmt after Close")
		return
	}

	if stmt = dc.m[query]; stmt != nil {
		return
	}
	if stmt, err = dc.ds.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("Prepare: %w", err)
		return
	}
	dc.m[query] = stmt

	return
}

func (dc *DBCache) Close() (err error) {
	ds, stmts := dc.getCloses()
	if ds != nil {
		if err = ds.Close(); err != nil {
			err = perrors.Errorf("dataSource.Close: %w", err)
		}
	}
	err = perrors.AppendError(err, dc.closeStmts(stmts))

	return
}

func (dc *DBCache) getStmts() (stmts []*sql.Stmt) {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	return dc.getStmts2()
}

func (dc *DBCache) getStmts2() (stmts []*sql.Stmt) {
	stmts = make([]*sql.Stmt, len(dc.m))
	i := 0
	for _, stmt := range dc.m {
		stmts[i] = stmt
		i++
	}
	dc.m = map[string]*sql.Stmt{}

	return
}

func (dc *DBCache) closeStmts(stmts []*sql.Stmt) (err error) {
	for _, stmt := range stmts {
		if e := stmt.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("stmt.Close: %w", err))
		}
	}

	return
}

func (dc *DBCache) getCloses() (ds parl.DataSource, stmts []*sql.Stmt) {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	if dc.ds == nil {
		return
	}

	ds = dc.ds
	dc.ds = nil
	stmts = dc.getStmts2()

	return
}
