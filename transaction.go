package dbx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Transaction begins or reuses a transaction, executes the provided operation,
// and handles commit or rollback automatically. If the context already contains
// a transaction, it will be reused unless the WithNewTransaction option is specified.
//
// Transaction lifecycle management:
//   - If a new transaction is created, it's committed on successful operation or rolled back on error.
//   - If an existing transaction is reused, commit/rollback is left to the outer transaction.
//   - Any panic during operation execution triggers rollback if a new transaction was created.
//   - If an operation and its rollback both fail, the returned error contains both failures.
//   - Transaction options apply only when a new transaction is created.
//
// Parameters:
//   - ctx: Parent Go context.
//   - beginner: A Beginner capable of creating transactions (typically a Database).
//   - op: Operation to execute within the transaction, taking a dbx.Context.
//   - opts: Optional configuration (e.g., isolation, read-only, always create new transaction).
//
// Returns:
//   - error: Any error from transaction creation, operation execution, or commit/rollback.
//
// Example:
//
//	err := dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
//	    _, err := txCtx.Executor().ExecContext(txCtx, "INSERT INTO users (name) VALUES (?)", "John")
//	    if err != nil { return err } // triggers automatic rollback
//	    _, err = txCtx.Executor().ExecContext(txCtx, "INSERT INTO profiles (user_id) VALUES (?)", userID)
//	    return err
//	})
func Transaction(ctx context.Context, beginner Beginner, op Operation, opts ...Option) error {
	_, err := transactionWithInternal(ctx, beginner, func(ctx Context) (interface{}, error) {
		return nil, op(ctx)
	}, opts)

	return err
}

// TransactionWithResult begins a transaction and executes an operation returning a typed result.
// Handles automatic commit/rollback and transaction reuse (see Transaction for rules).
//
// Parameters:
//   - ctx: Parent Go context.
//   - beginner: A Beginner capable of creating transactions (typically a Database).
//   - op: Operation to execute within the transaction that returns (T, error).
//   - setters: Optional configuration (transaction isolation, read-only, always create, etc.).
//
// Returns:
//   - T: Result returned by the operation (zero value if error).
//   - error: Any error from transaction creation, operation execution, or commit/rollback.
//
// Example:
//
//	userID, err := dbx.TransactionWithResult(ctx, db, func(txCtx dbx.Context) (int64, error) {
//	    var userID int64
//	    err := txCtx.Executor().QueryRowContext(
//	        txCtx,
//	        "INSERT INTO users (name) VALUES (?) RETURNING id",
//	        "John",
//	    ).Scan(&userID)
//	    return userID, err
//	})
func TransactionWithResult[T any](ctx context.Context, beginner Beginner, op OperationWithResult[T], setters ...Option) (T, error) {
	return transactionWithInternal(ctx, beginner, op, setters)
}

// transactionWithInternal contains core transaction logic for Transaction and TransactionWithResult.
//
// Transaction reuse/creation:
//   - By default, attempts to detect and reuse an existing transaction in context.
//   - If WithNewTransaction is specified or no transaction exists, creates a new one.
//
// Error handling and lifecycle:
//   - Rolls back on error or panic if a new transaction was created.
//   - Commits on success if a new transaction was created.
//   - Existing transactions are reused with lifecycle managed by caller.
//
// Parameters:
//   - ctx: Parent Go context.
//   - beginner: Capable of creating a new transaction.
//   - op: Operation to execute, returns (T, error).
//   - setters: List of functional options for transaction configuration.
//
// Returns:
//   - T: Operation result (zero value if error).
//   - error: Any error from transaction handling or op execution.
func transactionWithInternal[T any](ctx context.Context, beginner Beginner, op OperationWithResult[T], setters []Option) (T, error) {
	var zero T
	var tx Transactor
	var createdTx bool
	var dbCtx Context
	var operationReturned bool
	opts := newOptions(setters)

	if !opts.AlwaysCreate {
		// Reuse an existing transaction without requiring beginner to provide
		// unrelated context or execution capabilities.
		dbCtx = FromContext(ctx)
		if dbCtx != nil {
			if transactor, ok := dbCtx.Executor().(Transactor); ok {
				tx = transactor
			}
		}
	}

	if tx == nil {
		var err error
		createdTx = true

		// create a new transaction
		tx, err = beginner.BeginTx(ctx, opts.TxOptions)

		if err != nil {
			return zero, err
		}

		// create a new context with the transaction
		dbCtx = NewContext(ctx, tx)

		// Keep rollback armed until op returns so panics are cleaned up without
		// issuing another rollback after normal lifecycle handling.
		defer func() {
			if !operationReturned {
				_ = tx.Rollback() //nolint:errcheck // A panic must retain its original value.
			}
		}()
	}

	out, err := op(dbCtx)
	operationReturned = true

	if err != nil {
		if createdTx {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				return zero, errors.Join(err, fmt.Errorf("rollback transaction: %w", rollbackErr))
			}
		}

		return zero, err
	}

	if createdTx {
		if e := tx.Commit(); e != nil {
			return zero, e
		}
	}

	return out, nil
}
