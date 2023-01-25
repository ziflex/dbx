package dbx

import (
	"context"
	"database/sql"
)

type ctxKey struct{}

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
func Transaction(ctx context.Context, db Database, op Operation, opts ...Option) error {
	_, err := transactionWithInternal(ctx, db, func(ctx Context) (interface{}, error) {
		return nil, op(ctx)
	}, opts)

	return err
}

// TransactionWith begins a transaction with a given options, creates a context and passes the context to a given receiver
func TransactionWith[T any](ctx context.Context, db Database, op OperationWithResult[T], setters ...Option) (T, error) {
	return transactionWithInternal(ctx, db, op, setters)
}

func transactionWithInternal[T any](ctx context.Context, db Database, op OperationWithResult[T], setters []Option) (T, error) {
	var opts *sql.TxOptions

	if len(setters) > 0 {
		opts = new(sql.TxOptions)

		for _, setter := range setters {
			setter(opts)
		}
	}

	tx, err := db.BeginTx(ctx, opts)

	if err != nil {
		return *new(T), err
	}

	out, err := op(FromTx(ctx, tx))

	if err != nil {
		tx.Rollback()

		return *new(T), err
	}

	if e := tx.Commit(); e != nil {
		return *new(T), e
	}

	return out, nil
}
