package dbx

import (
	"context"
	"database/sql"
)

// Transaction begins or reuses a transaction, passes the context to a given receiver and handles the commit or rollback.
// Note: if the context is a transaction context, the transaction will be reused.
func Transaction(ctx context.Context, db Database, op Operation, opts ...Option) error {
	_, err := transactionWithInternal(FromOrNew(ctx, db), func(ctx Context) (interface{}, error) {
		return nil, op(ctx)
	}, opts)

	return err
}

// TransactionWith begins a transaction with a given options, creates a context and passes the context to a given receiver
func TransactionWith[T any](ctx context.Context, db Database, op OperationWithResult[T], setters ...Option) (T, error) {
	return transactionWithInternal(FromOrNew(ctx, db), op, setters)
}

func transactionWithInternal[T any](ctx Context, op OperationWithResult[T], setters []Option) (T, error) {
	dbCtx, ok := ctx.(*defaultContext)

	if !ok {
		return *new(T), ErrInvalidContext
	}

	db := dbCtx.db
	opts := newOptions(setters)

	var tx *sql.Tx
	var err error
	var createdTx bool

	if opts.AlwaysCreate {
		createdTx = true

		tx, err = db.BeginTx(ctx, opts.TxOptions)
	} else {
		tx = dbCtx.tx

		if tx == nil {
			createdTx = true

			tx, err = db.BeginTx(ctx, opts.TxOptions)
		}
	}

	if err != nil {
		return *new(T), err
	}

	out, err := op(FromTx(ctx, tx))

	if err != nil {
		if createdTx {
			tx.Rollback()
		}

		return *new(T), err
	}

	if createdTx {
		if e := tx.Commit(); e != nil {
			return *new(T), e
		}
	}

	return out, nil
}
