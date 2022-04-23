/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"
	"database/sql"
)

func (db *DB) QueryRow(query string, args ...any) (sqlRow *sql.Row, err error) {
	return db.DataSource().QueryRowContext(db.ctx, query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) (sqlRow *sql.Row, err error) {
	return db.DataSource().QueryRowContext(ctx, query, args...)
}

func (db *DB) QueryRowYear(year string, query string, args ...any) (sqlRow *sql.Row, err error) {
	return db.DataSource(year).QueryRowContext(db.ctx, query, args...)
}

func (db *DB) QueryRowYearContext(year string, ctx context.Context, query string, args ...any) (sqlRow *sql.Row, err error) {
	return db.DataSource(year).QueryRowContext(ctx, query, args...)
}
