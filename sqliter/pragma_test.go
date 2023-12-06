/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"context"
	"database/sql"
	"maps"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/psql"
)

func TestPragmaDB(t *testing.T) {
	var pragmas map[string]string
	var ctx = context.Background()
	var err error
	var schema = func(dataSource parl.DataSource, ctx context.Context) (err error) { return }
	var dbs = psql.DBFactory.NewDB(newMemDataSourceNamer(), schema)
	pragmas, err = PragmaDB(dbs, parl.NoPartition, ctx)
	if err != nil {
		panic(err)
	}
	t.Logf("pragmas: %v", pragmas)
	err = dbs.Close()
	if err != nil {
		panic(err)
	}
}

func TestPragma(t *testing.T) {
	//t.Fail()
	var exp = map[string]string{
		"foreignKeys": "0",
		"journal":     "memory",
		"timeout":     "0",
	}

	var pragmas map[string]string
	var ctx = context.Background()
	var err error
	var dataSource parl.DataSource

	dataSource, err = OpenDataSource(SQLiteMemoryDataSourceName)
	if err != nil {
		panic(err)
	}

	pragmas, err = Pragma(dataSource, ctx)
	if err != nil {
		t.Fatalf("Open err: %s", perrors.Short(err))
	}

	// pragmas: map[foreignKeys:0 journal:memory timeout:0]
	t.Logf("pragmas: %v", pragmas)

	if !maps.Equal(pragmas, exp) {
		t.Errorf("pragmas: %v exp %v", pragmas, exp)
	}
	err = dataSource.Close()
	if err != nil {
		panic(err)
	}
}

func TestPragmaSQL(t *testing.T) {
	//t.Fail()
	var exp = map[string]string{
		"foreignKeys": "0",
		"journal":     "memory",
		"timeout":     "0",
	}

	var pragmas map[string]string
	var ctx = context.Background()
	var err error

	// retrieve in-memory default pragmas
	var sqlDB *sql.DB
	if sqlDB, err = sql.Open(SQLiteDriverName, SQLiteMemoryDataSourceName); perrors.IsPF(&err, "Open %w", err) {
		return
	}

	pragmas, err = PragmaSQL(sqlDB, ctx)
	if err != nil {
		t.Fatalf("Open err: %s", perrors.Short(err))
	}

	// pragmas: map[foreignKeys:0 journal:memory timeout:0]
	t.Logf("pragmas: %v", pragmas)

	if !maps.Equal(pragmas, exp) {
		t.Errorf("pragmas: %v exp %v", pragmas, exp)
	}
	err = sqlDB.Close()
	if err != nil {
		panic(err)
	}
}

// memDataSourceNamer is like a sqliter.DataSourceNamer for in-memory databases
type memDataSourceNamer struct{ DataSourceNamer }

// newMemDataSourceNamer returns an in-memory datasource namer
func newMemDataSourceNamer() (dsnr parl.DataSourceNamer) { return &memDataSourceNamer{} }

// DSN returns the same in-memory database every time
func (n *memDataSourceNamer) DSN(...parl.DBPartition) (dsn parl.DataSourceName) {
	return parl.DataSourceName(SQLiteMemoryDataSourceName)
}
