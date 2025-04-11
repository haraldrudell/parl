/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"database/sql"
	"iter"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// NewResultSetIterator123 returns a go1.23 row-set iterator
//   - sqlRows: the result of a multiple-row query like SELECT
//   - scanFunc: scans a record into type T
//   - errp: error indicator
//   - —
//   - to release resources, rowIterator must be either:
//   - — read to end by a for range statement
//   - — be canceled by yield returning false
//   - — be part of a panic
//   - after iteration, errp should be inspected for errors from
//     SQL or otherwise
func NewResultSetIterator123[T any](sqlRows *sql.Rows, scanner parl.RowScanner[T], errp *error, fieldp ...*ResultSetIterator[T]) (rowIterator iter.Seq[T]) {
	parl.NilPanic("errp", errp)

	// get iterator storage
	var iterator *ResultSetIterator[T]
	if len(fieldp) > 0 {
		iterator = fieldp[0]
	}
	if iterator == nil {
		iterator = &ResultSetIterator[T]{}
	}

	*iterator = ResultSetIterator[T]{
		sqlRows: sqlRows,
		scanner: scanner,
		errp:    errp,
	}
	rowIterator = iterator.rowIterator
	return
}

// ResultSetIterator.rowIterator is [iter.Seq]
var _ iter.Seq[int] = (&ResultSetIterator[int]{}).rowIterator

func (i *ResultSetIterator[T]) rowIterator(yield func(row T) (keepGoing bool)) {

	// check for nil errors
	var err error
	if i.sqlRows == nil {
		err = parl.NilError("sqlRows")
	} else if i.scanner == nil {
		err = parl.NilError("scanner")
	}
	if err != nil {
		*i.errp = perrors.AppendError(*i.errp, err)
		return
	}
	defer parl.Close(i.sqlRows, i.errp)

	// iterate
	var t T
	for {

		// check for existwence of another row
		if !i.sqlRows.Next() {
			return // end of rows return
		}

		if t, err = i.scanner.Scan(i.sqlRows); err != nil {
			break
		} else if !yield(t) {
			return // yield ordered iteration to end return
		}
	}
	// iteration error

	*i.errp = perrors.AppendError(*i.errp, err)
}
