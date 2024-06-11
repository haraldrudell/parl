/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
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
	//t.Error("Logging on")
	var (
		// empty schema function
		schema = func(dataSource parl.DataSource, ctx context.Context) (err error) { return }
		exp    = map[string]string{
			"foreignKeys": "0",
			"journalMode": "memory",
			"busyTimeout": "0",
		}
	)

	var (
		pragmas   parl.Pragmas
		ctx       = context.Background()
		err       error
		dbMap     = psql.DBFactory.NewDB(newMemDataSourceNamer(), schema)
		pragmaMap map[string]string
	)

	// PragmaDB should match
	pragmas, err = PragmaDB(dbMap, parl.NoPartition, ctx)
	if err != nil {
		t.Fatalf("PragmaDB err %s", perrors.Short(err))
	}
	// pragmas: “busyTimeout: 0 foreignKeys: 0 journalMode: memory”
	t.Logf("pragmas: “%s”", pragmas)
	pragmaMap = pragmas.Map()
	if !maps.Equal(pragmaMap, exp) {
		t.Errorf("pragmas: %v exp %v", pragmaMap, exp)
	}

	// Close should succeed
	err = dbMap.Close()
	if err != nil {
		t.Fatalf("Close err %s", perrors.Short(err))
	}
}

func TestPragma(t *testing.T) {
	//t.Error("Logging on")
	var (
		exp = map[string]string{
			"foreignKeys": "0",
			"journalMode": "memory",
			"busyTimeout": "0",
		}
	)

	var (
		pragmas    parl.Pragmas
		pragmaMap  map[string]string
		ctx        = context.Background()
		err        error
		dataSource parl.DataSource
	)

	// create in-memory data source
	dataSource, err = OpenDataSource(SQLiteMemoryDataSourceName)
	if err != nil {
		panic(err)
	}

	// pragmas should match
	pragmas, err = Pragma(dataSource, ctx)
	if err != nil {
		t.Fatalf("Open err: %s", perrors.Short(err))
	}
	// pragmas: map[foreignKeys:0 journal:memory timeout:0]
	t.Logf("pragmas: %v", pragmas)
	pragmaMap = pragmas.Map()
	if !maps.Equal(pragmaMap, exp) {
		t.Errorf("pragmas: %v exp %v", pragmaMap, exp)
	}

	// Close should succeed
	err = dataSource.Close()
	if err != nil {
		panic(err)
	}
}

// TestPragmaSQL tests SQLite pragma using native [sql.DB] and [sql.Open] functions
func TestPragmaSQL(t *testing.T) {
	//t.Error("Logging on")
	var (
		exp = map[string]string{
			"foreignKeys": "0",
			"journalMode": "memory",
			"busyTimeout": "0",
		}
	)

	var (
		pragmas   parl.Pragmas
		pragmaMap map[string]string
		ctx       = context.Background()
		err       error
	)

	// create in-memory data source

	// native [sql.DB] object: lots of methods and fields
	var sqlDB *sql.DB
	sqlDB, err = sql.Open(SQLiteDriverName, SQLiteMemoryDataSourceName)
	if err != nil {
		t.Fatalf("sql.Open err %s", perrors.Short(err))
	}

	// pragmas should match
	pragmas, err = PragmaSQL(sqlDB, ctx)
	if err != nil {
		t.Fatalf("PragmaSQL err: %s", perrors.Short(err))
	}
	// pragmas: “busyTimeout: 0 foreignKeys: 0 journalMode: memory”
	t.Logf("pragmas: “%s”", pragmas)
	pragmaMap = pragmas.Map()
	if !maps.Equal(pragmaMap, exp) {
		t.Errorf("pragmas: %v exp %v", pragmaMap, exp)
	}

	// Close should succeed
	err = sqlDB.Close()
	if err != nil {
		t.Errorf("Close err %s", err)
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
