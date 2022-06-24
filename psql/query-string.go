/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"context"
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

func QueryString(label string, ctx context.Context, dataSource parl.DataSource,
	query string, args ...any) (value string, err error) {

	// prepare the sql statement
	var sqlStmt *sql.Stmt
	if sqlStmt, err = dataSource.PrepareContext(ctx, query); err != nil {
		err = perrors.Errorf("prepare %s: %w", label, err)
		return
	}
	defer func() {
		if e := sqlStmt.Close(); e != nil {
			err = perrors.AppendError(err, perrors.Errorf("close %s: %w", label, e))
		}
	}()

	// execute
	if value, err = ScanToString(sqlStmt.QueryRowContext(ctx, args...), nil); err != nil {
		err = perrors.Errorf("exec %s: %w", label, err)
		return
	}

	return
}
