/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"database/sql"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/iters"
)

// ScanFunc scans a record into type T
type ScanFunc[T any] func(sqlRows *sql.Rows) (t T, err error)

// ResultSetIterator is an iterator for a sql result-set
type ResultSetIterator[T any] struct {
	sqlRows  *sql.Rows
	scanFunc ScanFunc[T]
}

// NewResultSetIterator returns a result-set iterator for type T
func NewResultSetIterator[T any](sqlRows *sql.Rows, scanFunc ScanFunc[T]) (iterator iters.Iterator[T]) {
	if sqlRows == nil {
		parl.NilError("sqlRows")
	} else if scanFunc == nil {
		parl.NilError("scanFunc")
	}
	return iters.NewFunctionIterator((&ResultSetIterator[T]{
		sqlRows:  sqlRows,
		scanFunc: scanFunc,
	}).iteratorFunction)
}

// iteratorFunction handles cancel, error and end-of-records
func (i *ResultSetIterator[T]) iteratorFunction(isCancel bool) (t T, err error) {
	if isCancel {
		err = i.sqlRows.Close()
		return // cancel notification return
	} else if !i.sqlRows.Next() {
		err = parl.ErrEndCallbacks
		return // end of data return
	}
	return i.scanFunc(i.sqlRows)
}
