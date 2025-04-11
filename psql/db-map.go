/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
//   - [psql.DBMap] implements [parl.DB]
type DBMap struct {
	// dsnr is a SQL implementation-specific data source provider implementing:
	//	- possible partitioning
	//	- creation and query of databases
	dsnr parl.DataSourceNamer
	// schema executes application-specific SQL initialization for a new datasource
	//	- executes CREATE TABLE for tables and indexes
	//	- configures database-specific referential integrity and journaling
	schema func(dataSource parl.DataSource, ctx context.Context) (err error)
	// stateLock makes m thread-safe
	//	- makes Close and getOrCreateDBCache critical section
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

// QueryRow executes a query returning only its first row
//   - zero rows returns error: sql: no rows in result set: use [DB.Query]
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

// QueryString executes a query known to return zero or one row and first column a string value
//   - implemented by [sql.DB.QueryRowContext]
func (d *DBMap) QueryString(
	partition parl.DBPartition, query string, noRowsOk parl.NoRowsAction,
	ctx context.Context,
	args ...any,
) (value string, hasValue bool, err error) {

	// retrieve a possibly cached prepared statement
	var stmt psql2.Stmt
	if stmt, err = d.getStmt(partition, query, ctx); err != nil {
		return
	}

	if err = stmt.QueryRowContext(ctx, args...).Scan(&value); err != nil {
		if noRowsOk == parl.NoRowsOK && errors.Is(err, sql.ErrNoRows) {
			err = nil
			return
		}
		err = perrors.Errorf("QueryString.Scan: %w", err)
		return
	}
	hasValue = true

	return
}

// QueryInt executes a query known to return zero or one row and first column an int value
//   - implemented by [sql.DB.QueryRowContext]
func (d *DBMap) QueryInt(
	partition parl.DBPartition, query string, noRowsOk parl.NoRowsAction,
	ctx context.Context,
	args ...any,
) (value int, hasValue bool, err error) {

	// retrieve a possibly cached prepared statement
	var stmt psql2.Stmt
	if stmt, err = d.getStmt(partition, query, ctx); err != nil {
		return
	}

	if err = stmt.QueryRowContext(ctx, args...).Scan(&value); err != nil {
		if noRowsOk == parl.NoRowsOK && errors.Is(err, sql.ErrNoRows) {
			err = nil
			return
		}
		err = perrors.Errorf("QueryInt.Scan: %w", err)
		return
	}
	hasValue = true

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

// “psql.DBMap-parl.DB-0x14000316f60-#0-unclosed”
//   - uniquely identifies the [psql.DBMap] value
//   - how many mapped partitions
//   - close and error state
func (d *DBMap) String() (s string) {

	// “parlDB.String: psql.DBMap-parl.DB-0x14000062e28”
	//	- uniquely identifies the type
	//	- uniquely identifies the value
	s = fmt.Sprintf("%s-parl.DB-0x%x",
		dbMapTypeName,
		parl.Uintptr(d),
	)

	// if nil value
	if d == nil {
		return
	}

	var length, isInitialized = d.length()
	if !isInitialized {
		s += "-uninitialized"
		return
	}

	// “-#4-unclosed”
	var status string
	if errp := d.closeErr.Load(); errp == nil {
		status = "unclosed"
	} else if err := *errp; err == nil {
		status = "closed"
	} else {
		status = fmt.Sprintf("close-err:“%s”", perrors.Short(err))
	}
	s += fmt.Sprintf("-#%d-%s", length, status)

	return
}

func (d *DBMap) length() (length int, isInitialized bool) {
	d.stateLock.Lock()
	defer d.stateLock.Unlock()

	isInitialized = d.m != nil
	length = len(d.m)
	return
}

// getStmt obtains a cached statemnt or prepares the statement and caches it
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

// type name of [parl.DB] implementation: “psql.DBMap”
var dbMapTypeName = fmt.Sprintf("%T", DBMap{})
