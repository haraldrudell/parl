/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package psql

import (
	"context"
)

func (db *DB) Exec(query string, args ...any) (id int64, rows int64, err error) {
	return db.DataSource().ExecContext(db.ctx, query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (id int64, rows int64, err error) {
	return db.DataSource().ExecContext(ctx, query, args...)
}

func (db *DB) ExecYear(year string, query string, args ...any) (id int64, rows int64, err error) {
	return db.DataSource(year).ExecContext(db.ctx, query, args...)
}

func (db *DB) ExecYearContext(year string, ctx context.Context, query string, args ...any) (id int64, rows int64, err error) {
	return db.DataSource(year).ExecContext(ctx, query, args...)
}
