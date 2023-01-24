package dbx

import (
	"context"
	"database/sql"
)

type ctxKey struct{}

// With returns a new context with a given DB context.
func With(ctx context.Context, dbCtx Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, dbCtx)
}

// From returns a DB context from a given context.
func From(ctx context.Context) Context {
	if dbCtx, ok := ctx.Value(ctxKey{}).(Context); ok {
		return dbCtx
	}

	return nil
}

// Is returns true if the context is a DB context.
func Is(ctx context.Context) bool {
	_, ok := ctx.(Context)

	return ok
}

// As returns a DB context if the context is a DB context.
func As(ctx context.Context) (Context, bool) {
	c, ok := ctx.(Context)

	return c, ok
}

// Transaction begins a transaction, creates a context and passes the context to a given receiver
func Transaction[T any](ctx context.Context, db Database, op Operation[T]) (T, error) {
	return TransactionWith(ctx, db, op, nil)
}

// TransactionWith begins a transaction with a given options, creates a context and passes the context to a given receiver
func TransactionWith[T any](ctx context.Context, db Database, op Operation[T], opts *sql.TxOptions) (T, error) {
	tx, err := db.BeginTx(ctx, opts)

	if err != nil {
		return *new(T), err
	}

	out, err := op(WithTx(ctx, tx))

	if err != nil {
		tx.Rollback()

		return *new(T), err
	}

	if e := tx.Commit(); e != nil {
		return *new(T), e
	}

	return out, nil
}
