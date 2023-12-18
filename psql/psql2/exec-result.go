/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql2

import (
	"database/sql"
	"fmt"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// ExecResult makes sql.Result printable.
// sql.Result are obtained by invoking Exec* methods of sql.DB sql.Stmt sql.Conn sql.Tx.
// sql.Result is an interface with methods LastInsertId and RowsAffected.
// each driver provides an sql.Result implementation.
type ExecResult struct {
	ID   int64 // last insert ID, 0 if none like UPDATE
	rows int64 // number of rows affected
}

func NewExecResult(execResult sql.Result, e error) (result parl.ExecResult, err error) {
	r := ExecResult{}
	if err = e; err != nil {
		return
	}

	if r.ID, err = execResult.LastInsertId(); err != nil {
		err = perrors.Errorf("LastInsertId err: %w", err)
		return
	}
	if r.rows, err = execResult.RowsAffected(); err != nil {
		err = perrors.Errorf("RowsAffected err: %w", err)
		return
	}
	result = &r

	return
}

// ExecValues parses the result from an Exec* method into its values.
// if LastInsertId or RowsAffected fails, the error is added to err.
func ExecValues(execResult sql.Result, e error) (ID, rows int64, err error) {
	var result parl.ExecResult
	result, err = NewExecResult(execResult, e)
	r := result.(*ExecResult)
	ID = r.ID
	rows = r.rows

	return
}

// Get obtains last id and number of affected rows with errors separately
func (r *ExecResult) Get() (ID int64, rows int64) {
	ID = r.ID
	rows = r.rows
	return
}

func (r ExecResult) String() (s string) {
	return fmt.Sprintf("sql.Result: ID %d rows: %d", r.ID, r.rows)
}
