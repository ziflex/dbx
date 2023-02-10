package dbx

import (
	"context"
)

// Transaction begins or reuses a transaction, passes the context to a given receiver and handles the commit or rollback.
// Note: if the context is a transaction context, the transaction will be reused.
func Transaction(ctx context.Context, db Database, op Operation, opts ...Option) error {
	_, err := transactionWithInternal(ctx, db, func(ctx Context) (interface{}, error) {
		return nil, op(ctx)
	}, opts)

	return err
}

// TransactionWithResult begins a transaction with a given options, creates a context and passes the context to a given receiver
func TransactionWithResult[T any](ctx context.Context, db Database, op OperationWithResult[T], setters ...Option) (T, error) {
	return transactionWithInternal(ctx, db, op, setters)
}

func transactionWithInternal[T any](ctx context.Context, db Database, op OperationWithResult[T], setters []Option) (T, error) {
	var tx Transactor
	var createdTx bool
	var dbCtx Context
	opts := newOptions(setters)

	if !opts.AlwaysCreate {
		// retrieve existing or create a new context
		dbCtx = NewContextFrom(ctx, db)
		executor := dbCtx.Executor()

		// check if the executor is a transaction
		transactor, ok := executor.(Transactor)

		// if the executor is a transaction, use it
		if ok {
			tx = transactor
		}
	}

	if tx == nil {
		var err error
		createdTx = true

		// create a new transaction
		tx, err = db.BeginTx(ctx, opts.TxOptions)

		if err != nil {
			return *new(T), err
		}

		// create a new context with the transaction
		dbCtx = NewContext(ctx, tx)
	}

	out, err := op(dbCtx)

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
