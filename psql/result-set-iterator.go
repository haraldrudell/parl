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
//   - invoked after [sql.Rows.Next] returned true
//   - intended to execute [sql.Rows.Scan] and return result
type ScanFunc[T any] func(sqlRows *sql.Rows) (t T, err error)

// ResultSetIterator is an iterator for a sql result-set
type ResultSetIterator[T any] struct {
	sqlRows  *sql.Rows
	scanFunc ScanFunc[T]
}

// NewResultSetIterator returns a result-set iterator for type T
//   - scanFunc invokes [sql.Rows.Scan] and returns result
//   - to return multiple values, use a tuple value-container similar to
//     [github.com/haraldrudell/parl/pfs.ResultEntry]
//   - note that scanFunc is self-contained.
//     Providing a method value as scanFunc causes allocation 20 ns M1 Max.
//     Providing a top-level function is allocation-free.
//
// Usage:
//
//	var sqlRows *sql.Rows
//	if sqlRows, err = o.Query(parl.NoPartition, myQuery, o.ctx); perrors.IsPF(&err, "query %w", err) {
//	  return
//	}
//	iterator = psql.NewResultSetIterator(sqlRows, scanFunc)
//	defer iterator.Cancel(&err)
//	for item, _ := iterator.Init(); iterator.Cond(&item); {
//	  …
//	func scanFunc(sqlRows *sql.Rows) (item Item, err error) {
//	  err = sqlRows.Scan(&item)
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
