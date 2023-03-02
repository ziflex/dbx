package dbx

import (
	"context"
	"database/sql"
	"io"
)

type (
	// Executor provides an abstraction for sql.DB and sql.Tx
	Executor interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
		ExecContext(dbContext context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(dbContext context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(dbContext context.Context, query string, args ...interface{}) *sql.Row
	}

	// Transactor provides an abstraction for sql.Tx
	Transactor interface {
		Commit() error
		Rollback() error

		Executor
	}

	// Beginner provides an abstraction for sql.DB
	Beginner interface {
		Begin() (*sql.Tx, error)
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	}

	// ContextCreator provides an executor context creation.
	ContextCreator interface {
		// Context creates a new executor context
		Context(ctx context.Context) Context
	}

	// Operation is a user-defined database operation that needs to be performed within a transaction
	Operation func(ctx Context) error

	// OperationWithResult is a user-defined database operation that needs to be performed within a transaction and returns a result.
	OperationWithResult[T any] func(ctx Context) (T, error)

	// Database interface represents an entry point for the context
	Database interface {
		io.Closer
		ContextCreator
		Beginner
		Executor
	}

	// Context provides a general purpose abstraction to communication between domain services and data repositories.
	Context interface {
		context.Context

		// Executor returns a sql executor.
		// If transaction provided, sql.Tx will be returned, otherwise sql.DB.
		Executor() Executor
	}
)
