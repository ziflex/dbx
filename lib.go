package dbx

import (
	"context"
	"database/sql"
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

	// ContextReceiver is a context receiver.
	ContextReceiver func(ctx Context) error

	// ContextCreator provides a db context creation.
	ContextCreator interface {
		// Context creates a new db context
		Context(ctx context.Context) Context
	}

	// Operation is a user-defined database operation that needs to be performed within a transaction
	Operation func(executor Executor) error

	// Database interface represents an entry point for the context
	Database interface {
		Beginner
		ContextCreator

		// Transaction begins a transaction, creates a context and passes the context to a given receiver
		Transaction(ctx context.Context, receiver ContextReceiver) error

		// TransactionWith begins a transaction with a given options, creates a context and passes the context to a given receiver
		TransactionWith(ctx context.Context, receiver ContextReceiver, opts *sql.TxOptions) error
	}

	// Context provides a general purpose abstraction to communication between domain services and data repositories.
	Context interface {
		context.Context

		// Executor returns a sql executor.
		// If transaction provided, sql.Tx will be returned, otherwise sql.DB.
		Executor() Executor
	}
)
