package dbcontext

import (
	"context"
	"database/sql"
)

type (
	// Receiver is a context receiver
	Receiver func(ctx Context) error

	// Database interface represents an entry point for the context
	Database interface {
		// Transaction begins a transaction, creates a context and passes the context to a given receiver
		Transaction(ctx context.Context, receiver Receiver) error

		// Transaction begins a transaction with a given options, creates a context and passes the context to a given receiver
		TransactionWith(ctx context.Context, receiver Receiver, opts *sql.TxOptions) error

		// Context creates a new context
		Context(ctx context.Context) Context
	}

	// Executor provides an abstraction for sql.DB and sql.Tx
	Executor interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
		ExecContext(dbContext context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(dbContext context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(dbContext context.Context, query string, args ...interface{}) *sql.Row
	}

	// Operation is a user-defined database operation that needs to be performed within a transaction
	Operation func(executor Executor) error

	// Context provides a general purpose abstraction to communication between domain services and data repositories.
	Context interface {
		context.Context

		// Returns an sql executor.
		// If transaction provided, sql.Tx will be returned, otherwise sql.DB.
		Executor() Executor

		// Begin executes a given operation within a transaction.
		// If the context was given a transaction, the transaction will be used, otherwise new one will be created.
		Begin(operation Operation) error

		// Begin executes a given operation within a transaction with passed transaction options.
		// If the context was given a transaction, the transaction will be used and the options will be ignored,
		// otherwise new one will be created.
		BeginWith(operation Operation, opts *sql.TxOptions) error
	}
)
