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
	ds       parl.DataSource
	lock     sync.Mutex
	m        map[string]*sql.Stmt // behind lock
	closeErr error                // behind lock
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
		return // state error exit
	}

	if stmt = dc.m[query]; stmt != nil {
		return // found cached statement exit
	}

	if stmt, err = dc.ds.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("Prepare: %w", err)
		return // statement prepare failed exit
	}
	dc.m[query] = stmt

	return // new cached statement exit
}

func (dc *DBCache) WrapStmt(stmt *sql.Stmt) (stm Stmt) {
	if stmtWrapper, ok := dc.ds.(StmtWrapper); ok {
		return stmtWrapper.WrapStmt(stmt)
	}
	return stmt
}

func (dc *DBCache) Close() (err error) {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	ds := dc.ds
	if ds == nil {
		return dc.closeErr // already closed exit
	}

	// close data source
	dc.ds = nil // flag object closed
	if err = ds.Close(); err != nil {
		err = perrors.Errorf("dataSource.Close: %w", err)
	}

	// close cached statements
	m := dc.m
	dc.m = nil // drop stmt references
	for _, stmt := range m {
		if e := stmt.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("stmt.Close: %w", err))
		}
	}
	if err != nil {
		dc.closeErr = err // indicate close error status
	}

	return
}
