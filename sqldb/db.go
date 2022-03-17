/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package sqldb interfaces database/sql
package sqldb

import (
	"context"
	"database/sql"

	"github.com/haraldrudell/parl"
)

const (
	contextDBformat = "db.db"
)

// StoreDB saves db handle in context
func StoreDB(ctx context.Context, db *sql.DB) {
	parl.StoreInContext(ctx, contextDBformat, db)
}

// GetDB obtains db handle from context
func GetDB(ctx context.Context) (db *sql.DB) {
	db, ok := ctx.Value(contextDBformat).(*sql.DB)
	if !ok {
		panic(parl.New("No db in context"))
	}
	return
}

// DiscardDB removes db handle from context
func DiscardDB(ctx context.Context) {
	parl.DelFromContext(ctx, contextDBformat)
}
