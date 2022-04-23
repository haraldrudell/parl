/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"context"
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	_ "modernc.org/sqlite"
)

const (
	sqLiteDriverName = "sqlite"
)

type DataSource struct {
	DB *sql.DB
	DBUtil
}

// NewDB get a DB object that repreents the databases in a directory
func NewDataSource(dataSourceName string) (dataSource parl.DataSource, err error) {
	d := DataSource{}

	if d.DB, err = sql.Open(sqLiteDriverName, dataSourceName); err != nil {
		err = perrors.Errorf("sql.Open(%s %s): %w", sqLiteDriverName, dataSourceName, err)
		return
	}

	dataSource = &d
	return
}

func (dbFile *DataSource) ExecContext(ctx context.Context, query string, args ...any) (id int64, rows int64, err error) {

	var sqlStmt *sql.Stmt
	if sqlStmt, err = dbFile.DB.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("PrepareContext: %w", err)
		return
	}

	var sqlResult sql.Result
	if sqlResult, err = sqlStmt.ExecContext(ctx, args...); err != nil {
		err = perrors.Errorf("ExecContext: %w", err)
		return
	}

	if id, err = sqlResult.LastInsertId(); err != nil {
		err = perrors.Errorf("LastInsertId: %w", err)
		return
	}

	if rows, err = sqlResult.RowsAffected(); err != nil {
		err = perrors.Errorf("RowsAffected: %w", err)
		return
	}

	return
}

func (dbFile *DataSource) QueryContext(ctx context.Context,
	cb func(sqlRows *sql.Rows) (err error),
	query string, args ...any) (err error) {

	var sqlStmt *sql.Stmt
	if sqlStmt, err = dbFile.DB.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("PrepareContext: %w", err)
		return
	}

	var sqlRows *sql.Rows
	if sqlRows, err = sqlStmt.QueryContext(ctx, args...); err != nil {
		err = perrors.Errorf("QueryContext: %w", err)
		return
	}
	defer func() {
		if e := sqlRows.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("sqlRows.Close: %w", err))
		}
	}()

	err = cb(sqlRows)

	return
}

func (dbFile *DataSource) QueryRowContext(ctx context.Context, query string, args ...any) (sqlRow *sql.Row, err error) {

	var sqlStmt *sql.Stmt
	if sqlStmt, err = dbFile.DB.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("PrepareContext: %w", err)
		return
	}

	sqlRow = sqlStmt.QueryRowContext(ctx, args...)

	return
}

func (df *DataSource) Close() (err error) {
	err = df.DB.Close()

	return
}
