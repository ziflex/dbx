package dbx

import (
	"context"
)

// Transaction begins or reuses a transaction, executes the provided operation,
// and handles commit or rollback automatically. If the context already contains
// a transaction, it will be reused unless the WithNewTransaction option is specified.
//
// The transaction lifecycle is managed automatically:
//   - If a new transaction is created, it will be committed on success or rolled back on error
//   - If an existing transaction is reused, commit/rollback is left to the outer transaction
//   - Any panic during operation execution will trigger a rollback if a new transaction was created
//
// Parameters:
//   - ctx: The parent Go context
//   - db: Database instance to create transaction from (if needed)
//   - op: Operation to execute within the transaction
//   - opts: Optional transaction configuration (isolation level, read-only, etc.)
//
// Returns:
//   - error: Any error from transaction creation, operation execution, or commit/rollback
//
// Example:
//
//	err := dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
//	    _, err := txCtx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
//	    if err != nil {
//	        return err // This will trigger automatic rollback
//	    }
//	    _, err = txCtx.Executor().Exec("INSERT INTO profiles (user_id) VALUES (?)", userID)
//	    return err
//	})
func Transaction(ctx context.Context, db Database, op Operation, opts ...Option) error {
	_, err := transactionWithInternal(ctx, db, func(ctx Context) (interface{}, error) {
		return nil, op(ctx)
	}, opts)

	return err
}

// TransactionWithResult begins a transaction and executes an operation that returns a typed result.
// Like Transaction, it handles automatic commit/rollback and transaction reuse, but allows
// the operation to return a value along with any error.
//
// The transaction lifecycle follows the same rules as Transaction:
//   - New transactions are committed on success or rolled back on error
//   - Existing transactions are reused and their lifecycle managed by the outer scope
//
// Parameters:
//   - ctx: The parent Go context
//   - db: Database instance to create transaction from (if needed)
//   - op: Operation to execute that returns a typed result
//   - setters: Optional transaction configuration options
//
// Returns:
//   - T: The result returned by the operation (zero value if error occurred)
//   - error: Any error from transaction creation, operation execution, or commit/rollback
//
// Example:
//
//	userID, err := dbx.TransactionWithResult(ctx, db, func(txCtx dbx.Context) (int64, error) {
//	    result, err := txCtx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
//	    if err != nil {
//	        return 0, err
//	    }
//	    return result.LastInsertId()
//	})
func TransactionWithResult[T any](ctx context.Context, db Database, op OperationWithResult[T], setters ...Option) (T, error) {
	return transactionWithInternal(ctx, db, op, setters)
}

// transactionWithInternal implements the core transaction logic used by both
// Transaction and TransactionWithResult functions. It handles transaction creation,
// reuse detection, operation execution, and automatic commit/rollback.
//
// Transaction Reuse Logic:
//   - If opts.AlwaysCreate is false (default), checks for existing transaction in context
//   - If existing transaction found, reuses it and delegates lifecycle management to outer scope
//   - If no existing transaction or AlwaysCreate is true, creates a new transaction
//
// Error Handling:
//   - If operation returns an error and a new transaction was created, automatically rolls back
//   - If operation succeeds and a new transaction was created, automatically commits
//   - If reusing existing transaction, no automatic commit/rollback occurs
//
// Parameters:
//   - ctx: Parent Go context
//   - db: Database instance for creating new transactions
//   - op: Operation to execute within transaction scope
//   - setters: Transaction configuration options
//
// Returns:
//   - T: Result from the operation (zero value if error occurred)
//   - error: Any error from transaction management or operation execution
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
