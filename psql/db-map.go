/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package psql provides cached, shared, partitioned database objects with
// cached prepared statements cancelable by context.
//
//   - [NewDBMap] provides cached, shared database implementation objects.
//     database access is using cached prepared statements and
//     access using application name and partition name like year.
//   - [NewResultSetIterator] provides a Go for-statements abstract result-set iterator
//   - [ScanFunc] is the signature for preparing custom result-set iterators
//   - seamless statement-retry, remedying concurrency-deficient databases such as
//     SQLite3
//   - —
//   - [TrimSql] trims SQL statements in Go raw string literals, multi-line strings enclosed by
//     the back-tick ‘`’ character
//   - [ColumnType] describes columns of a result-set
//   - convenience method for single-value queries: [DBMap.QueryInt] [DBMap.QueryString]
//   - [SqlExec] pprovides statement execution prior to obtaining a cached database, ie. for
//     seamlessly preparing the schema
package psql

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/psql/psql2"
)

// DBMap provides:
//   - caching of SQL-implementation-specific database objects
//   - a cache for prepared statements via methods
//     Exec Query QueryRow QueryString QueryInt
type DBMap struct {
	// dsnr is a SQL implementation-specific data source provider implementing:
	//	- possible partitioning
	//	- creation and query of databases
	dsnr parl.DataSourceNamer
	// schema executes application-specific SQL initialization for a new datasource
	//	- executes CREATE of tables and indexes
	//	- configures database-specific referential integrity and journaling
	schema func(dataSource parl.DataSource, ctx context.Context) (err error)

	stateLock sync.Mutex
	m         map[parl.DataSourceName]*psql2.StatementCache // behind stateLock
	closeErr  atomic.Pointer[error]                         // written behind stateLock
}

// NewDBMap returns a database connection and prepared statement cache
// for dsnr that implements [parl.DB]
func NewDBMap(
	dsnr parl.DataSourceNamer,
	schema func(dataSource parl.DataSource, ctx context.Context) (err error),
) (dbMap *DBMap) {
	return &DBMap{
		dsnr:   dsnr,
		schema: schema,
		m:      make(map[parl.DataSourceName]*psql2.StatementCache),
	}
}

func NewDBMap2(
	dsnr parl.DataSourceNamer,
	schema func(dataSource parl.DataSource, ctx context.Context) (err error),
	getp *func(dataSourceName parl.DataSourceName,
		ctx context.Context) (dbStatementCache *psql2.StatementCache, err error),
) (dbMap *DBMap) {
	dbMap = NewDBMap(dsnr, schema)
	*getp = dbMap.getOrCreateDBCache
	return
}

// Exec executes a query not returning any rows
func (d *DBMap) Exec(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (execResult parl.ExecResult, err error) {
	var stmt psql2.Stmt
	if stmt, err = d.getStmt(partition, query, ctx); err != nil {
		return
	}
	if execResult, err = psql2.NewExecResult(stmt.ExecContext(ctx, args...)); err != nil {
		err = perrors.Errorf("Exec: %w", err)
		return
	}

	return
}

// Query executes a query returning zero or more rows
func (d *DBMap) Query(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (sqlRows *sql.Rows, err error) {
	var stmt psql2.Stmt
	if stmt, err = d.getStmt(partition, query, ctx); err != nil {
		return
	}
	if sqlRows, err = stmt.QueryContext(ctx, args...); err != nil {
		err = perrors.Errorf("Query: %w", err)
		return
	}

	return
}

// Query executes a query known to return exactly one row
func (d *DBMap) QueryRow(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (sqlRow *sql.Row, err error) {
	var stmt psql2.Stmt
	if stmt, err = d.getStmt(partition, query, ctx); err != nil {
		return
	}
	sqlRow = stmt.QueryRowContext(ctx, args...)
	if err = sqlRow.Err(); err != nil {
		err = perrors.Errorf("QueryRow: %w", err)
		return
	}

	return
}

// Query executes a query known to return exactly one row and returns its string value
func (d *DBMap) QueryString(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (value string, err error) {
	var stmt psql2.Stmt
	if stmt, err = d.getStmt(partition, query, ctx); err != nil {
		return
	}
	if err = stmt.QueryRowContext(ctx, args...).Scan(&value); err != nil {
		err = perrors.Errorf("QueryString.Scan: %w", err)
		return
	}

	return
}

// Query executes a query known to return exactly one row and returns its int value
func (d *DBMap) QueryInt(
	partition parl.DBPartition, query string, ctx context.Context,
	args ...any) (value int, err error) {
	var stmt psql2.Stmt
	if stmt, err = d.getStmt(partition, query, ctx); err != nil {
		return
	}
	if err = stmt.QueryRowContext(ctx, args...).Scan(&value); err != nil {
		err = perrors.Errorf("QueryInt.Scan: %w", err)
		return
	}

	return
}

// Close shuts down the statement cache and the data source
func (d *DBMap) Close() (err error) {

	// close check outside lock
	if ep := d.closeErr.Load(); ep != nil {
		err = *ep
		return // return obtained close result
	}
	d.stateLock.Lock()
	defer d.stateLock.Unlock()

	// close check inside lock
	if ep := d.closeErr.Load(); ep != nil {
		err = *ep
		return // another thread already closed
	}

	// close dbCache objects
	var dbCache = d.m
	d.m = nil // drop dbCache references
	for _, db := range dbCache {
		if e := db.Close(); e != nil {
			err = perrors.AppendError(err, e)
		}
	}

	d.closeErr.Store(&err) // store close status

	return
}

// getStmt obtains a cached statemrnt or prepares the statement and caches it
func (d *DBMap) getStmt(
	partition parl.DBPartition, query string, ctx context.Context,
) (stmt psql2.Stmt, err error) {

	// obtain the statement cache
	var dbCache *psql2.StatementCache
	if dbCache, err = d.getOrCreateDBCache(d.dsnr.DSN(partition), ctx); err != nil {
		return // closed or failure return
	}

	// obtain the statement
	var sqlStmt *sql.Stmt
	if sqlStmt, err = dbCache.Stmt(query, ctx); err != nil {
		return // closed or failure returrn
	}
	// possibly wrap the statement
	stmt = dbCache.WrapStmt(sqlStmt)

	return
}

// getOrCreateDBCache returns a cached database object or
// creates, caches and returns a database object
func (d *DBMap) getOrCreateDBCache(dataSourceName parl.DataSourceName,
	ctx context.Context) (dbStatementCache *psql2.StatementCache, err error) {

	// status check outside lock
	if ep := d.closeErr.Load(); ep != nil {
		err = perrors.NewPF("invocation after parl.DB close")
		return // bad status exit
	}
	d.stateLock.Lock()
	defer d.stateLock.Unlock()

	// status check inside lock
	if ep := d.closeErr.Load(); ep != nil {
		err = perrors.NewPF("invocation after parl.DB close")
		return // bad status exit
	}

	// try cache
	if dbStatementCache = d.m[dataSourceName]; dbStatementCache != nil {
		return // cached DB object exit
	}

	// create dataSource for new dbCache instance
	var dataSource parl.DataSource
	if dataSource, err = d.dsnr.DataSource(dataSourceName); err != nil {
		return // datasource create failure exit
	}
	defer d.getEnd(&err, dataSourceName, &dataSource, &dbStatementCache)

	// initialize schema
	if err = d.schema(dataSource, ctx); err != nil {
		return // schema failure exit
	}
	dbStatementCache = psql2.NewStatementCache(dataSource)

	return // good exit
}

// getEnd handles success or failure in creating statementCache
func (d *DBMap) getEnd(
	errp *error,
	dataSourceName parl.DataSourceName,
	dataSourcep *parl.DataSource,
	dbStatementCachep **psql2.StatementCache,
) {

	// success case
	if *errp == nil {
		d.m[dataSourceName] = *dbStatementCachep // success: store new object
		return
	}

	// error case
	//	- dataSource is present
	//	- dbStatementCache is not present
	if e := (*dataSourcep).Close(); perrors.IsPF(&e, "dataSource.Close: %w", e) {
		*errp = perrors.AppendError(*errp, e)
	}
}
