/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// StatementCache caches prepared statements for a data source
type StatementCache struct {
	// the datasource containing SQL tables
	DataSource parl.DataSource

	mLock sync.Mutex
	//	- key: SQL statement
	//	- value: a cached prepared statement
	//	- behind mlock
	m        map[string]*sql.Stmt
	closeErr atomic.Pointer[error] // written behind mLock
}

// NewStatementCache returns a cache for prepared statements for a data source
func NewStatementCache(dataSource parl.DataSource) (cache *StatementCache) {
	return &StatementCache{
		DataSource: dataSource,
		m:          make(map[string]*sql.Stmt),
	}
}

// Stmt returns a cached prepared statement from the cache or
// prepares, caches and returns a new prepared statement
func (c *StatementCache) Stmt(query string, ctx context.Context) (stmt *sql.Stmt, err error) {

	// close check outside lock
	if c.DataSource == nil {
		err = perrors.NewPF("Stmt after Close")
		return // state error exit
	}
	c.mLock.Lock()
	defer c.mLock.Unlock()

	// close check inside lock
	if c.DataSource == nil {
		err = perrors.NewPF("Stmt after Close")
		return // state error exit
	}

	// try cache
	if stmt = c.m[query]; stmt != nil {
		return // found cached statement exit
	}

	// prepare and cache a new statement
	if stmt, err = c.DataSource.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("Prepare: %w", err)
		return // statement prepare failed exit
	}
	c.m[query] = stmt

	return // new cached statement exit
}

// WrapStmt retruns a wrapped statement if the data source support it
//   - wrapper is used for retries of databases like SQLite3
//     that may always return busy errors
func (c *StatementCache) WrapStmt(stmt *sql.Stmt) (stm Stmt) {

	// if wrapper supported, return a wrapped statement
	if stmtWrapper, ok := c.DataSource.(StmtWrapper); ok {
		stm = stmtWrapper.WrapStmt(stmt)
		return
	}

	// return unwrapped statement
	return stmt
}

// Close shuts down the statement cache
func (c *StatementCache) Close() (err error) {

	// close check outside lock
	if ep := c.closeErr.Load(); ep != nil {
		err = *ep
		return // already closed
	}
	c.mLock.Lock()
	defer c.mLock.Unlock()

	// close check inside lock
	if ep := c.closeErr.Load(); ep != nil {
		err = *ep
		return // another thread already closed
	}

	// close data source
	if err = c.DataSource.Close(); err != nil {
		err = perrors.Errorf("dataSource.Close: %w", err)
	}

	// close cached statements
	var statements = c.m
	c.m = nil // drop stmt references
	for _, stmt := range statements {
		if e := stmt.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("stmt.Close: %w", err))
		}
	}

	c.closeErr.Store(&err)

	return
}
