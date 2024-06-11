/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/psql/psql2"
)

// pragmaList is a sorted list of the pragma names
// that the database is queried for
var pragmaList = []string{
	"foreignKeys", "journalMode", "busyTimeout",
}

// pragmaStatements translates an index in pragmaList to
// the SQL statement retrieving that pragma value
var pragmaStatements = map[string]string{
	pragmaList[0]: "PRAGMA foreign_keys",
	pragmaList[1]: "PRAGMA journal_mode",
	pragmaList[2]: "PRAGMA busy_timeout",
}

// sqliterPragma is a private implementation of [parl.Pragmas]
type sqliterPragma map[string]string

// Pragma returns SQLite3 pragma database-settings based on a data source
//   - parl.DataSource is a caching of prepared statements obtained from
//     OpenDataSource
//   - When Pragma is used to retieve database pragmas, the issued queries will be cached
//     in dataSource
func Pragma(dataSource parl.DataSource, ctx context.Context) (pragmas parl.Pragmas, err error) {

	// retrieve a map of pragma values
	var pragmaMap = make(map[string]string, len(pragmaList))
	for _, key := range pragmaList {
		var sqlRows *sql.Rows
		if sqlRows, err = psql2.Query(key, ctx, dataSource, pragmaStatements[key]); err != nil {
			return
		} else if err = scanPragma(sqlRows, key, pragmaMap); err != nil {
			return
		}
	}

	// return *sqliterPragma that implements [parl.Pragmas]
	var p sqliterPragma = pragmaMap
	pragmas = &p

	return
}

// Pragma returns some common SQLite3 database settings
//   - parl.DB is a map of multiple DB objects facilitating partitioning
//   - in normal use, [parl.DB] is not available from [psql.NewDBMap].
//     The DB value is cached internally and retrieved via its data source name
//   - When PragmaDB is used to retieve database pragmas, the issued queries will be cached
//     in the [parl.DB] data source
func PragmaDB(db parl.DB, partition parl.DBPartition, ctx context.Context) (pragmas parl.Pragmas, err error) {

	// retrieve a map of pragma values
	var pragmaMap = make(map[string]string, len(pragmaList))
	for _, key := range pragmaList {
		// cannot use QueryString because zero rows may be returned
		//	- cannot use QueryRow because it expects exactly one row
		var sqlRows *sql.Rows
		if sqlRows, err = db.Query(partition, pragmaStatements[key], ctx); err != nil {
			return
		} else if err = scanPragma(sqlRows, key, pragmaMap); err != nil {
			return
		}
	}

	// return *sqliterPragma that implements [parl.Pragmas]
	var pragmaValue sqliterPragma = pragmaMap
	pragmas = &pragmaValue

	return
}

// PragmaSQL returns common SQLite3 database settings
//   - PragmaSQL uses native [sql.DB] and [sql.DB.QueryContext] functions
//   - sqlDB is obtained from [sql.Open]
//   - — sqlDB does not cache the queries issued by PragmaSQL
//   - in-memory database [parl.Pragmas.String]: map[foreignKeys:0 journal:memory timeout:0]
func PragmaSQL(sqlDB *sql.DB, ctx context.Context) (pragmas parl.Pragmas, err error) {

	// retrieve a map of pragma values
	var pragmaMap = make(map[string]string, len(pragmaList))
	for _, key := range pragmaList {
		var sqlRows *sql.Rows
		if sqlRows, err = sqlDB.QueryContext(ctx, pragmaStatements[key]); err != nil {
			return
		} else if err = scanPragma(sqlRows, key, pragmaMap); err != nil {
			return
		}
	}

	// return *sqliterPragma that implements [parl.Pragmas]
	var pragmaValue sqliterPragma = pragmaMap
	pragmas = &pragmaValue

	return
}

// Map returns pragmas in the form of a map
func (p *sqliterPragma) Map() (pragmaMap map[string]string) { return *p }

func (p *sqliterPragma) String() (s string) {
	// p is *map[string]string
	if p == nil || *p == nil {
		return "<nil>"
	}
	var pragmaList = make([]string, len(*p))
	var i int
	for pragma, value := range *p {
		pragmaList[i] = fmt.Sprintf("%s: %s", pragma, value)
		i++
	}
	slices.Sort(pragmaList)
	s = strings.Join(pragmaList, "\x20")
	return
}

// scanPragma scans sqlRows adding any value to pragmaMap
func scanPragma(sqlRows *sql.Rows, key string, pragmaMap map[string]string) (err error) {
	defer parl.Close(sqlRows, &err)

	if sqlRows.Next() {
		var value string
		if err = sqlRows.Scan(&value); perrors.IsPF(&err, "Scan %w", err) {
			return // scan error return
		} else if sqlRows.Next() {
			err = perrors.NewPF("PRAGMA returned more than one row")
			return // more than one row error return
		}
		pragmaMap[key] = value
	}

	return
}
