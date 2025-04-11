/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"iter"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

func TestNewResultSetIterator123(t *testing.T) {
	const (
		value         = 1
		yieldCountOne = 1
	)
	var (
		iterationPanic = errors.New("panic")
	)

	var (
		sqlRows, sqlRowsNil   *sql.Rows
		scannerNil            parl.RowScanner[int]
		errpNil               *error
		panic, actualErr, err error
		mockDB                *mockSQL
		intScannr             *intScanner
		yields                *yielder
		forRangeCounter       int
	)

	// func NewResultSetIterator123[T any](
	// sqlRows *sql.Rows, scanner parl.RowScanner[T],
	// errp *error, fieldp ...*ResultSetIterator[T])
	// (rowIterator iter.Seq[T])
	//	- type Seq[V any] func(yield func(V) bool)
	var rowIterator iter.Seq[int]

	// errp nil should panic
	rowIterator, panic = invokeNew(sqlRowsNil, scannerNil, errpNil)
	_ = rowIterator
	if panic == nil {
		t.Error("FAIL missing panic: NewResultSetIterator123 errp nil")
	}

	// first iteration should return row
	mockDB = newMockSQL(value)
	sqlRows = mockDB.sqlRows()
	actualErr = nil
	intScannr = newIntScanner()
	yields = newYielder()
	rowIterator = NewResultSetIterator123(sqlRows, intScannr, &actualErr)
	rowIterator(yields.yieldTrue)
	if actualErr != nil {
		t.Errorf("FAIL iteration error %s", actualErr)
	}
	if yields.count != yieldCountOne {
		t.Errorf("FAIL yield function not invoked")
	}
	if yields.value != value {
		t.Errorf("FAIL yield value bad %d exp %d", yields.value, value)
	}

	// second iteration should end
	rowIterator(yields.yieldTrue)
	if actualErr != nil {
		t.Errorf("FAIL iteration error %s", actualErr)
	}
	if yields.count > yieldCountOne {
		t.Errorf("FAIL yield function unexpectedly invoked")
	}
	err = mockDB.expectations()
	if err != nil {
		t.Errorf("FAIL SQL expectations err: “%s”", err)
	}

	// yield returning false should cancel iteration
	mockDB = newMockSQL(value)
	sqlRows = mockDB.sqlRows()
	actualErr = nil
	intScannr = newIntScanner()
	yields = newYielder()
	rowIterator = NewResultSetIterator123(sqlRows, intScannr, &actualErr)
	rowIterator(yields.yieldFalse)
	if actualErr != nil {
		t.Errorf("FAIL iteration error %s", actualErr)
	}
	if yields.count != yieldCountOne {
		t.Errorf("FAIL yield function not invoked")
	}
	if yields.value != value {
		t.Errorf("FAIL yield value bad %d exp %d", yields.value, value)
	}
	err = mockDB.expectations()
	if err != nil {
		t.Errorf("FAIL SQL expectations err: “%s”", err)
	}

	// for range iteration should work
	mockDB = newMockSQL(value)
	sqlRows = mockDB.sqlRows()
	actualErr = nil
	intScannr = newIntScanner()
	forRangeCounter = 0
	for v := range NewResultSetIterator123(sqlRows, intScannr, &actualErr) {
		if v != value {
			t.Errorf("FAIL iteration value bad %d exp %d", v, value)
		}
		forRangeCounter++
	}
	if forRangeCounter != yieldCountOne {
		t.Errorf("FAIL bad iteration count %d exp %d", forRangeCounter, yieldCountOne)
	}
	if actualErr != nil {
		t.Errorf("FAIL iteration error %s", actualErr)
	}
	err = mockDB.expectations()
	if err != nil {
		t.Errorf("FAIL SQL expectations err: “%s”", err)
	}

	// for range iteration panic should work
	mockDB = newMockSQL(value)
	sqlRows = mockDB.sqlRows()
	actualErr = nil
	intScannr = newIntScanner()
	rowIterator = NewResultSetIterator123(sqlRows, intScannr, &actualErr)
	panic = panicDuringIteration(rowIterator, iterationPanic)
	if panic == nil {
		t.Error("FAIL missing panic")
	} else if !errors.Is(panic, iterationPanic) {
		t.Errorf("FAIL bad panic during iteration : “%s”", panic)
	}
	if actualErr != nil {
		t.Errorf("FAIL iteration error %s", actualErr)
	}
	err = mockDB.expectations()
	if err != nil {
		t.Errorf("FAIL SQL expectations err: “%s”", err)
	}
}

// invokeNew captures panic when invoking NewResultSetIterator123
func invokeNew(sqlRows *sql.Rows, scanner parl.RowScanner[int], errp *error) (rowIterator iter.Seq[int], panic error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &panic)

	rowIterator = NewResultSetIterator123(sqlRows, scanner, errp)
	return
}

// mockSQL provides sql.Rows with mocked content
type mockSQL struct {
	// mockDB is mock implementation of [sql.DB]
	//	- created by [go-sqlmock.New]
	//	- can execute queries providing mock [sql.Rows]
	mockDB *sql.DB
	// sqlMock is an orchestrating object for mockDB
	sqlMock sqlmock.Sqlmock
}

// newMockSQL returns generator of [sql.Rows] with mocked content
//   - values: column-values returned in a single SQL query
//   - m: a mock database generating fake result set
//   - [mockDB.sqlRows] returns mocked [sql.Rows]
//   - e
func newMockSQL(values ...any) (m *mockSQL) {
	m = &mockSQL{}

	// create the mock database
	var err error
	if m.mockDB, m.sqlMock, err = sqlmock.New(); perrors.Is(&err, "sqlmock.New %w", err) {
		panic(err)
	}

	// prepare the query
	// column names are “col1”…
	var colNames = make([]string, len(values))
	for i := range colNames {
		colNames[i] = fmt.Sprintf("col%d", i+1)
	}

	// create a mock query returning [sql.Rows]
	var driverValues = make([]driver.Value, len(values))
	for i, anyValue := range values {
		driverValues[i] = anyValue
	}
	var expectedQuery = m.sqlMock.ExpectQuery(queryName).WillReturnRows(
		m.sqlMock.NewRows(colNames).
			AddRow(driverValues...),
	)
	// expect rows.SQL to be closed
	expectedQuery.RowsWillBeClosed()

	// expect Close
	// m.sqlMock.ExpectClose() //.WillReturnError(fmt.Errorf(`ops`))

	return
}

// sqlRows returns sql.Rows with single-row values content
//   - sqlRows: single-row mock [sql.Rows]
//   - similar to a SELECT statement
//   - values go into columns
func (m *mockSQL) sqlRows() (sqlRows *sql.Rows) {

	// execute the query
	var err error
	if sqlRows, err = m.mockDB.Query(queryName); perrors.Is(&err, "mockDb.Query %w", err) {
		panic(err)
	}
	return
}

// expectations returns werror if expected events did not take place
func (m *mockSQL) expectations() (err error) {
	err = m.sqlMock.ExpectationsWereMet()
	return
}

// intScanner is [parl.RowScanner] for single-column int
type intScanner struct{}

// intScanner is [parl.RowScanner]
var _ parl.RowScanner[int] = &intScanner{}

// newIntScanner returns a [parl.RowScanner] for single-column int
func newIntScanner() (i *intScanner) { return &intScanner{} }

// Scan scans a row from sqlRows
func (i *intScanner) Scan(sqlRows *sql.Rows) (t int, err error) {
	err = sqlRows.Scan(&t)

	return
}

// yielder provides yield functions for iteration testing
type yielder struct {
	value int
	count int
}

// newYielder provides yield functions for iteration testing
func newYielder() (y *yielder) { return &yielder{} }

// yieldTrue is int yield function returning true
func (y *yielder) yieldTrue(value int) (keepGoing bool) {
	y.value = value
	y.count++
	return true
}

// yieldFalse is int yield function returning false
func (y *yielder) yieldFalse(value int) (keepGoing bool) {
	y.value = value
	y.count++
	return false
}

// iterationPanic tests panic during iteration
func panicDuringIteration(rowIterator iter.Seq[int], iterationPanic error) (errPanic error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &errPanic)

	for v := range rowIterator {
		_ = v
		panic(iterationPanic)
	}

	return
}
