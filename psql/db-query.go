/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"
	"database/sql"
)

func (db *DB) Query(
	cb func(sqlRows *sql.Rows) (err error),
	query string, args ...any) (err error) {
	return db.DataSource().QueryContext(db.ctx, cb, query, args...)
}

func (db *DB) QueryContext(ctx context.Context,
	cb func(sqlRows *sql.Rows) (err error),
	query string, args ...any) (err error) {
	return db.DataSource().QueryContext(ctx, cb, query, args...)
}

func (db *DB) QueryYear(year string,
	cb func(sqlRows *sql.Rows) (err error),
	query string, args ...any) (err error) {
	return db.DataSource(year).QueryContext(db.ctx, cb, query, args...)
}

func (db *DB) QueryYearContext(
	year string, ctx context.Context,
	cb func(sqlRows *sql.Rows) (err error),
	query string, args ...any) (err error) {
	return db.DataSource(year).QueryContext(ctx, cb, query, args...)
}
