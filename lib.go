// Package dbx provides a context-driven database abstraction layer for Go.
//
// It wraps the standard database/sql package to provide better context management,
// automatic transaction handling, and a unified interface for database operations.
// The package follows Go best practices for context propagation and makes it easier
// to work with transactions while maintaining compatibility with existing database/sql code.
//
// Key Features:
//   - Context-driven design that embeds database connections within Go contexts
//   - Automatic transaction lifecycle management with support for nested transactions
//   - Unified interface that works the same for both direct DB operations and transactions
//   - Interface-based design for maximum flexibility and testability
//   - Zero magic - predictable behavior with no hidden surprises
//   - Minimal overhead - thin layer that doesn't compromise performance
//
// Basic Usage:
//
//	db, err := sql.Open("postgres", connectionString)
//	if err != nil {
//	    return err
//	}
//	defer db.Close()
//
//	// Create a dbx database instance
//	dbxDB := dbx.New(db)
//	defer dbxDB.Close()
//
//	// Use context for database operations
//	ctx := context.Background()
//	dbCtx := dbxDB.Context(ctx)
//
//	// Execute queries
//	rows, err := dbCtx.Executor().Query("SELECT * FROM users")
//
// Transaction Example:
//
//	err := dbx.Transaction(ctx, dbxDB, func(txCtx dbx.Context) error {
//	    _, err := txCtx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
//	    return err
//	})
package dbx

import (
	"context"
	"database/sql"
	"io"
)

type (
	// Executor provides an abstraction for both sql.DB and sql.Tx, allowing
	// uniform database operations regardless of whether you're working with
	// a direct database connection or within a transaction.
	//
	// This interface mirrors the core methods available in both sql.DB and sql.Tx,
	// providing both context-aware and legacy methods for maximum compatibility.
	Executor interface {
		// Exec executes a query without returning any rows.
		// The args are for any placeholder parameters in the query.
		Exec(query string, args ...interface{}) (sql.Result, error)

		// Query executes a query that returns rows, typically a SELECT.
		// The args are for any placeholder parameters in the query.
		Query(query string, args ...interface{}) (*sql.Rows, error)

		// QueryRow executes a query that is expected to return at most one row.
		// QueryRow always returns a non-nil value. Errors are deferred until
		// Row's Scan method is called.
		QueryRow(query string, args ...interface{}) *sql.Row

		// ExecContext executes a query without returning any rows.
		// The args are for any placeholder parameters in the query.
		ExecContext(dbContext context.Context, query string, args ...interface{}) (sql.Result, error)

		// QueryContext executes a query that returns rows, typically a SELECT.
		// The args are for any placeholder parameters in the query.
		QueryContext(dbContext context.Context, query string, args ...interface{}) (*sql.Rows, error)

		// QueryRowContext executes a query that is expected to return at most one row.
		// QueryRowContext always returns a non-nil value. Errors are deferred until
		// Row's Scan method is called.
		QueryRowContext(dbContext context.Context, query string, args ...interface{}) *sql.Row
	}

	// Transactor provides an abstraction for sql.Tx, extending Executor with
	// transaction control methods. This interface allows managing transaction
	// lifecycle through explicit commit and rollback operations.
	Transactor interface {
		// Commit commits the transaction. If the transaction has already been
		// committed or rolled back, Commit returns an error.
		Commit() error

		// Rollback aborts the transaction. If the transaction has already been
		// committed or rolled back, Rollback returns an error.
		Rollback() error

		// Embed Executor to provide all database operation methods
		Executor
	}

	// Beginner provides an abstraction for sql.DB's transaction creation methods.
	// This interface allows starting new transactions with different options and
	// isolation levels.
	Beginner interface {
		// Begin starts a transaction with default options. The default isolation level
		// is dependent on the driver.
		Begin() (*sql.Tx, error)

		// BeginTx starts a transaction with the provided context and options.
		// The provided context is used until the transaction is committed or rolled back.
		// If the context is canceled, the sql package will roll back the transaction.
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	}

	// ContextCreator provides the ability to create dbx Context instances
	// from standard Go contexts. This is typically implemented by Database
	// to bootstrap the context-driven database operations.
	ContextCreator interface {
		// Context creates a new dbx Context from a standard Go context.
		// The returned Context embeds the database executor and can be used
		// for database operations within the provided context's lifecycle.
		Context(ctx context.Context) Context
	}

	// Operation represents a user-defined database operation that needs to be
	// performed within a transaction. Operations receive a dbx Context and
	// should return an error if the operation fails.
	//
	// Example:
	//	op := func(ctx dbx.Context) error {
	//		_, err := ctx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
	//		return err
	//	}
	Operation func(ctx Context) error

	// OperationWithResult represents a user-defined database operation that needs to be
	// performed within a transaction and returns a typed result. This is useful for
	// operations that need to return data, such as inserted IDs or query results.
	//
	// Example:
	//	op := func(ctx dbx.Context) (int64, error) {
	//		result, err := ctx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
	//		if err != nil {
	//			return 0, err
	//		}
	//		return result.LastInsertId()
	//	}
	OperationWithResult[T any] func(ctx Context) (T, error)

	// Database interface represents the main entry point for dbx operations.
	// It combines database connection management, context creation, transaction
	// initiation, and direct query execution capabilities.
	//
	// Database implementations should wrap sql.DB and provide context-aware
	// database operations while maintaining compatibility with the standard
	// database/sql interface.
	Database interface {
		// Embed io.Closer to allow proper database connection cleanup
		io.Closer

		// Embed ContextCreator to bootstrap dbx contexts
		ContextCreator

		// Embed Beginner to support transaction creation
		Beginner

		// Embed Executor to support direct database operations
		Executor
	}

	// Context provides a context-aware abstraction for database operations.
	// It extends the standard Go context.Context while embedding a database
	// executor, allowing seamless propagation of both context information
	// and database connection/transaction state.
	//
	// Context instances automatically handle the appropriate executor (sql.DB or sql.Tx)
	// based on whether they were created from a direct database connection or
	// within a transaction scope.
	Context interface {
		// Embed standard Go context for deadline, cancellation, and value propagation
		context.Context

		// Executor returns the appropriate sql executor for this context.
		// Returns sql.Tx if this context was created within a transaction,
		// otherwise returns sql.DB for direct database operations.
		Executor() Executor
	}
)
