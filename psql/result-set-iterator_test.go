/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
)

func TestResultSetIterator(t *testing.T) {
	var expValue = 1

	var value, zeroValue int
	var hasValue bool
	var err error

	// mock SQL database
	//	- m does not have to be closed
	var m = newMockDB()

	var iterator iters.Iterator[int]
	var reset = func() {
		var sqlRows = m.sqlRows(expValue)
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

// mockDB provides sql.Rows with mocked content
type mockDB struct {
	mockDb  *sql.DB
	sqlMock sqlmock.Sqlmock
}

// newMockDB returns generator of sql.Rows with mocked content
func newMockDB() (m *mockDB) {
	m = &mockDB{}
	var err error
	if m.mockDb, m.sqlMock, err = sqlmock.New(); perrors.Is(&err, "sqlmock.New %w", err) {
		panic(err)
	}
	return
}

// sqlRows returns sql.Rows with single-row values content
func (m *mockDB) sqlRows(values ...any) (sqlRows *sql.Rows) {
	var colNames = make([]string, len(values))
	for i := range colNames {
		colNames[i] = fmt.Sprintf("col%d", i+1)
	}
	var driverValues = make([]driver.Value, len(values))
	for i, anyValue := range values {
		driverValues[i] = anyValue
	}
	var expectedQuery = m.sqlMock.ExpectQuery(queryName).WillReturnRows(
		m.sqlMock.NewRows(colNames).
			AddRow(driverValues...),
	)
	_ = expectedQuery
	var err error
	if sqlRows, err = m.mockDb.Query(queryName); perrors.Is(&err, "mockDb.Query %w", err) {
		panic(err)
	}
	return
}
