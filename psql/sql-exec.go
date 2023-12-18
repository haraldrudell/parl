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
	"github.com/haraldrudell/parl/psql/psql2"
)

// SqlExec is used when parl.DB is not available, for example to implement the schema
// function provided to DBFactory.NewDB().
// query is an sql statement that does not return any rows.
// label is used in error messages.
// SqlExec uses parl.DataSource obtained from DataSourceNamer.DataSource().
// Because the cached prepared statements and the partitioning of parl.DB are not
// available, SqlExec uses any sql.DB method.
func SqlExec(label string, ctx context.Context, dataSource parl.DataSource,
	query string, args ...any) (err error) {

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
	var execResult parl.ExecResult
	if execResult, err = psql2.NewExecResult(sqlStmt.ExecContext(ctx, args...)); err != nil {
		err = perrors.Errorf("exec %s: %w", label, err)
		return
	}

	parl.Debug("%s result: %s", label, execResult)

	return
}
