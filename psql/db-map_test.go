/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/sqliter"
)

func TestDBMap(t *testing.T) {
	//t.Error("Logging on")
	var (
		inMemoryDatasourceNamer = sqliter.MemDsnr
		emptySchema             = func(dataSource parl.DataSource, ctx context.Context) (err error) { return }
		expNilString            = "psql.DBMap-parl.DB-0x0"
		expUninitializedSuffix  = "-uninitialized"
		expContains             = "-#"
	)

	var (
		s                  string
		nilDBMap           *DBMap
		uninitializedDBMap = &DBMap{}
	)

	// Close() Exec() Query() QueryInt() QueryRow()
	// QueryString() String()
	var parlDB *DBMap = NewDBMap(inMemoryDatasourceNamer, emptySchema)

	// String for nil should be 0x0
	s = nilDBMap.String()
	// nil parlDB.String: psql.DBMap-parl.DB-0x0
	t.Logf("nil parlDB.String: %s", s)
	if s != expNilString {
		t.Errorf("nil String: %q exp %q", s, expNilString)
	}

	// String for uninitialized should match
	s = uninitializedDBMap.String()
	// uninitialized parlDB.String: psql.DBMap-parl.DB-0x14000340f00-uninitialized
	t.Logf("uninitialized parlDB.String: %s", s)
	if !strings.HasSuffix(s, expUninitializedSuffix) {
		t.Errorf("unitizialized String Suffix: %q exp %q", s, expUninitializedSuffix)
	}

	// String for initialized should match
	s = parlDB.String()
	// “parlDB.String: psql.DBMap-parl.DB-0x14000340f60-#0-unclosed”
	t.Logf("parlDB.String: %s", s)
	if !strings.Contains(s, expContains) {
		t.Errorf("String not Contains %q exp %q", s, expContains)
	}
}

func TestDBMapQueryString(t *testing.T) {
	//t.Error("Logging on")
	var (
		// selects a SQLite3 shared in-memory database
		inMemoryDatasourceNamer = sqliter.MemDsnr
		// SQLite3 empty schema
		emptySchema = func(dataSource parl.DataSource, ctx context.Context) (err error) { return }
		// a query returning at least one row and first column string
		queryManyRows = `PRAGMA foreign_keys`
		// a query returning zero rows
		queryNoRows = `PRAGMA abc`
	)

	var (
		ctx      = context.Background()
		value    string
		hasValue bool
		err      error
	)

	// Close() Exec() Query() QueryInt() QueryRow()
	// QueryString() String()
	var parlDB *DBMap = NewDBMap(inMemoryDatasourceNamer, emptySchema)

	// query returning one result
	value, hasValue, err = parlDB.QueryString(parl.NoPartition, queryManyRows, parl.NoRowsError, ctx)
	if err != nil {
		t.Errorf("QueryString err “%s”", err)
	}
	if !hasValue {
		t.Error("hasValue false")
	}
	if value == "" {
		t.Error("value empty")
	}

	// query returning no rows returning error
	value, hasValue, err = parlDB.QueryString(parl.NoPartition, queryNoRows, parl.NoRowsError, ctx)
	if err == nil {
		t.Errorf("QueryString missing error")
	} else {
		// err: *errorglue.errorStack *fmt.wrapError *errors.errorString “QueryString.Scan: sql: no rows in result set’
		t.Logf("err: %s “%s’", errorglue.DumpChain(err), err)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("QueryString bad error “%s”", err)
		}
	}
	if hasValue {
		t.Error("hasValue true")
	}
	if value != "" {
		t.Error("value not empty")
	}

	// query returning no rows returning error
	value, hasValue, err = parlDB.QueryString(parl.NoPartition, queryNoRows, parl.NoRowsOK, ctx)
	if err != nil {
		t.Errorf("QueryString bad error “%s”", err)
	}
	if hasValue {
		t.Error("hasValue true")
	}
	if value != "" {
		t.Error("value not empty")
	}
}
