/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import (
	"database/sql"
	"fmt"
)

func ColumnType(typ *sql.ColumnType) (s string) {
	if typ == nil {
		return "ColumnType:nil"
	}
	var nullableS string
	if nullable, ok := typ.Nullable(); ok {
		nullableS = fmt.Sprintf("%t", nullable)
	} else {
		nullableS = "?"
	}
	var lengthS string
	if length, ok := typ.Length(); ok {
		lengthS = fmt.Sprintf("%d", length)
	} else {
		lengthS = "?"
	}
	var decimalSizeS string
	if precision, scale, ok := typ.DecimalSize(); ok {
		decimalSizeS = fmt.Sprintf("precision %d scale %d", precision, scale)
	} else {
		decimalSizeS = "?"
	}
	return fmt.Sprintf("name: %q database type-name: %s reflect-type: %s"+
		" nullable %s length: %s decimal-size: %s",
		typ.Name(), typ.DatabaseTypeName(), typ.ScanType(),
		nullableS, lengthS, decimalSizeS,
	)
}
