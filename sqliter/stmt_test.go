/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

const (
	driverName = "sqlite"
	filename   = "test.db"
)

func TestExecContext(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	query := "PRAGMA journal_mode"

	var sqlDB *sql.DB
	var err error
	var sqlStmt *sql.Stmt
	var sqlResult sql.Result
	var lastInsertId int64
	var rowsAffected int64

	if sqlDB, err = sql.Open(driverName, filepath.Join(dir, filename)); err != nil {
		t.Errorf("sql.Open: %v", err)
		t.FailNow()
	}

	// sqlDB: *sql.DB
	//t.Logf("sqlDB: %T", sqlDB)

	if sqlStmt, err = sqlDB.PrepareContext(ctx, query); err != nil {
		t.Errorf("PrepareContext: %v", err)
		t.FailNow()
	}

	// sqlStmt: *sql.Stmt
	//t.Logf("sqlStmt: %T", sqlStmt)

	if sqlResult, err = sqlStmt.ExecContext(ctx); err != nil {
		t.Errorf("ExecContext: %v", err)
	}

	// sql.driverResult
	// sql is some driver-related package
	//t.Logf("sqlResult: %T", sqlResult)

	if lastInsertId, err = sqlResult.LastInsertId(); err != nil {
		t.Errorf("ExecContext: %v", err)
	}
	if rowsAffected, err = sqlResult.RowsAffected(); err != nil {
		t.Errorf("RowsAffected: %v", err)
	}

	// LastInsertId: 0 RowsAffected: 0
	t.Logf("LastInsertId: %d RowsAffected: %d",
		lastInsertId,
		rowsAffected,
	)

	if err = sqlStmt.Close(); err != nil {
		t.Errorf("sqlStmt.Close: %v", err)
	}
	if err = sqlDB.Close(); err != nil {
		t.Errorf("sqlDB.Close: %v", err)
	}

	//t.Fail()
}

func TestQueryContext(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	query := "PRAGMA journal_mode"

	var sqlDB *sql.DB
	var err error
	var sqlStmt *sql.Stmt
	var sqlRows *sql.Rows

	if sqlDB, err = sql.Open(driverName, filepath.Join(dir, filename)); err != nil {
		t.Errorf("sql.Open: %v", err)
		t.FailNow()
	}

	if sqlStmt, err = sqlDB.PrepareContext(ctx, query); err != nil {
		t.Errorf("PrepareContext: %v", err)
		t.FailNow()
	}

	if sqlRows, err = sqlStmt.QueryContext(ctx); err != nil {
		t.Errorf("ExecContext: %v", err)
	}

	// sqlRows: *sql.Rows
	//t.Logf("sqlRows: %T", sqlRows)

	for sqlRows.Next() {
		if sqlRows.Scan() != nil {
			break
		}
	}
	if err = sqlRows.Err(); err != nil {
		t.Errorf("Scan: %v", err)
	}

	if err = sqlRows.Close(); err != nil {
		t.Errorf("sqlRows.Close: %v", err)
	}
	if err = sqlStmt.Close(); err != nil {
		t.Errorf("sqlStmt.Close: %v", err)
	}
	if err = sqlDB.Close(); err != nil {
		t.Errorf("sqlDB.Close: %v", err)
	}

	//t.Fail()
}

func TestQueryRowContext(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	query := "PRAGMA journal_mode"

	var sqlDB *sql.DB
	var err error
	var sqlStmt *sql.Stmt
	var sqlRow *sql.Row

	if sqlDB, err = sql.Open(driverName, filepath.Join(dir, filename)); err != nil {
		t.Errorf("sql.Open: %v", err)
		t.FailNow()
	}

	if sqlStmt, err = sqlDB.PrepareContext(ctx, query); err != nil {
		t.Errorf("PrepareContext: %v", err)
		t.FailNow()
	}

	if sqlRow = sqlStmt.QueryRowContext(ctx); err != nil {
		t.Errorf("ExecContext: %v", err)
	}

	// sqlRow: *sql.Row
	//t.Logf("sqlRow: %T", sqlRow)

	var s string
	if err = sqlRow.Scan(&s); err != nil {
		t.Errorf("sqlRow.Scan: %v", err)
	}

	// value: "delete"
	t.Logf("value: %q", s)

	if err = sqlStmt.Close(); err != nil {
		t.Errorf("sqlStmt.Close: %v", err)
	}
	if err = sqlDB.Close(); err != nil {
		t.Errorf("sqlDB.Close: %v", err)
	}

	//t.Fail()
}
