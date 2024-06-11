/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql2

import (
	"context"
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Query issues a query by preparing a statement for dataSource
func Query(label string, ctx context.Context, dataSource parl.DataSource,
	query string, args ...any) (sqlRows *sql.Rows, err error) {

	// prepare the sql statement
	var sqlStmt *sql.Stmt
	if sqlStmt, err = dataSource.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("prepare %s: %w", label, err)
		return
	}
	defer closeStmt(sqlStmt, label, &err)

	// execute
	if sqlRows, err = sqlStmt.QueryContext(ctx, args...); err != nil {
		err = perrors.Errorf("exec %s: %w", label, err)
	}

	return
}

func closeStmt(sqlStmt *sql.Stmt, label string, errp *error) {
	var err = sqlStmt.Close()
	if err == nil {
		return
	}
	*errp = perrors.AppendError(*errp, perrors.Errorf("close %s: %w", label, err))
}
