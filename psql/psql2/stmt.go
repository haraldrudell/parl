/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql2

import (
	"context"
	"database/sql"
)

type Stmt interface {
	ExecContext(ctx context.Context, args ...any) (sqlResult sql.Result, err error)
	QueryContext(ctx context.Context, args ...any) (sqlRows *sql.Rows, err error)
	QueryRowContext(ctx context.Context, args ...any) (sqlRow *sql.Row)
}

type StmtWrapper interface {
	WrapStmt(stmt *sql.Stmt) (stm Stmt)
}
