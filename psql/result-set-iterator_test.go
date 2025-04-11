/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
)

func TestResultSetIterator(t *testing.T) {
	const (
		expValue = 1
	)

	var (
		value, zeroValue int
		hasValue         bool
		err              error
		// mock SQL database
		//	- m does not have to be closed
		m *mockSQL
	)

	var iterator iters.Iterator[int]
	var reset = func() {
		m = newMockSQL(expValue)
		var sqlRows = m.sqlRows()
		var scanFunc = newIteratorConverter(sqlRows, expValue).scanFunc
		iterator = NewResultSetIterator(sqlRows, scanFunc)
	}

	// Cancel should return no error
	reset()
	err = iterator.Cancel()
	if err != nil {
		t.Errorf("Cancel err: %s", perrors.Short(err))
	}

	// Next returns value
	reset()
	value, hasValue = iterator.Next()
	if value != expValue {
		t.Errorf("Next value %d exp %d", value, expValue)
	}
	if !hasValue {
		t.Error("Next hasValue false")
	}
	err = iterator.Cancel()
	if err != nil {
		t.Errorf("Next err: %s", perrors.Short(err))
	}

	// NextNext returns no value
	reset()
	value, hasValue = iterator.Next()
	_ = value
	_ = hasValue
	value, hasValue = iterator.Next()
	if value != zeroValue {
		t.Errorf("Next2 value %d exp %d", value, zeroValue)
	}
	if hasValue {
		t.Error("Next2 hasValue true")
	}
	err = iterator.Cancel()
	if err != nil {
		t.Errorf("NextNext err: %s", perrors.Short(err))
	}
}

// iteratorConverter provides a simpleConverter for sql iterator
type iteratorConverter struct {
	sqlRows *sql.Rows
	value   int
}

// newIteratorConverter returns a simpleConverter object for sql iterator
func newIteratorConverter(sqlRows *sql.Rows, value int) (c *iteratorConverter) {
	return &iteratorConverter{sqlRows: sqlRows, value: value}
}

// func(sqlRows *sql.Rows) (t T, err error)
var _ ScanFunc[int]

// scanFunc is an sql iterator simple converter for type int
func (i *iteratorConverter) scanFunc(sqlRows *sql.Rows) (value int, err error) {
	if sqlRows != i.sqlRows {
		panic(errors.New("scanFunc sqlRows bad"))
	}
	value = i.value
	i.value = 0
	return
}

// mockSQL query name
const queryName = "x"
