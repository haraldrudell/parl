/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"context"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/psql"
)

var pragmaList = []string{
	"foreignKeys", "journal", "timeout",
}

var pragmaMap = map[string]string{
	pragmaList[0]: "PRAGMA foreign_keys",
	pragmaList[1]: "PRAGMA journal_mode",
	pragmaList[2]: "PRAGMA busy_timeout",
}

// Pragma returns some common SQLite3 database settings
func Pragma(dataSource parl.DataSource, ctx context.Context) (pragmas map[string]string, err error) {

	pragmas = make(map[string]string)
	var value string
	for _, key := range pragmaList {
		if value, err = psql.QueryString(key, ctx, dataSource, pragmaMap[key]); err != nil {
			return
		}
		pragmas[key] = value
	}

	return
}

// Pragma returns some common SQLite3 database settings
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
