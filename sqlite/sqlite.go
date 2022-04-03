/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package sqlite wraps modernc.org/sqlite
package sqlite

import (
	"path"

	// Database driver

	"github.com/haraldrudell/parl/pos"
	_ "modernc.org/sqlite"
)

const (
	dotLocalDir = ".local"
	shareDir    = "share"
	dbExt       = ".db"
	// DriverNameSqlite is name of SQLite driver to database/sql
	DriverNameSqlite = "sqlite"
)

// DsnSqlite builds a dsn identifying a database
func DsnSqlite(appName string, dir string) (dataSourceName string) {
	if dir == "" {
		dir = pos.HomeDir(path.Join(dotLocalDir, shareDir, appName))
	}
	dataSourceName = path.Join(dir, appName+dbExt)
	return
}
