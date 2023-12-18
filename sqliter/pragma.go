/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"context"
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/psql/psql2"
)

var pragmaList = []string{
	"foreignKeys", "journalMode", "busyTimeout",
}

var pragmaMap = map[string]string{
	pragmaList[0]: "PRAGMA foreign_keys",
	pragmaList[1]: "PRAGMA journal_mode",
	pragmaList[2]: "PRAGMA busy_timeout",
}

// Pragma returns some common SQLite3 database settings
//   - parl.DataSource is a caching of prepared statements obatined from
//     OpenDataSource
func Pragma(dataSource parl.DataSource, ctx context.Context) (pragmas map[string]string, err error) {

	pragmas = make(map[string]string)
	var value string
	for _, key := range pragmaList {
		if value, err = psql2.QueryString(key, ctx, dataSource, pragmaMap[key]); err != nil {
			return
		}
		pragmas[key] = value
	}

	return
}

// Pragma returns some common SQLite3 database settings
//   - parl.DB is a map of multiple DB objects facilitating partitioning
//   - in normal use parl.DB is not available from psql.NewDBMap
func PragmaDB(db parl.DB, partition parl.DBPartition, ctx context.Context) (pragmas map[string]string, err error) {

	pragmas = make(map[string]string)
	var value string
	for _, key := range pragmaList {
		if value, err = db.QueryString(partition, pragmaMap[key], ctx); err != nil {
			return
		}
		pragmas[key] = value
	}

	return
}

// Pragma returns some common SQLite3 database settings
//   - sqlDB is obtained from [sql.Open]
//   - in-memory returns: map[foreignKeys:0 journal:memory timeout:0]
func PragmaSQL(sqlDB *sql.DB, ctx context.Context) (pragmas map[string]string, err error) {

	pragmas = make(map[string]string)
	var result string
	var sqlRow *sql.Row
	for _, key := range pragmaList {
		sqlRow = sqlDB.QueryRowContext(ctx, pragmaMap[key])
		if err = sqlRow.Err(); perrors.IsPF(&err, "Query %s: %w", key, err) {
			return
		} else if sqlRow.Scan(&result); perrors.IsPF(&err, "Query %s: %w", key, err) {
			return
		}
		pragmas[key] = result
	}

	return
}
