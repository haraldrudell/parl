/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package psql augments database/sql
package psql

import (
	"context"
	"database/sql"

	"github.com/haraldrudell/parl/perrors"
)

const (
	contextDBformat = "db.db"
)

// StoreDB saves db handle in context
func StoreDB(ctx context.Context, db *sql.DB) {
	StoreInContext(ctx, contextDBformat, db)
}

// GetDB obtains db handle from context
func GetDB(ctx context.Context) (db *sql.DB) {
	db, ok := ctx.Value(contextDBformat).(*sql.DB)
	if !ok {
		panic(perrors.New("No db in context"))
	}
	return
}

// DiscardDB removes db handle from context
func DiscardDB(ctx context.Context) {
	DelFromContext(ctx, contextDBformat)
}
