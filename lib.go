package dbcontext

import (
	"context"
	"database/sql"
	"time"
)

type (
	// Database interface a represents an entry point for the context
	Database interface {
		Context(ctx context.Context) Context
	}

	// Executor provides an abstraction of *sql.DB and *sql.Tx
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
		// If the context was given a transaction, the transaction will be used, otherwise new will be created.
		Begin(operation Operation) error

		// Begin executes a given operation within a transaction with passed transaction options.
		// If the context was given a transaction, the transaction will be used and the options will be ignored, otherwise new will be created.
		BeginWith(operation Operation, opts *sql.TxOptions) error
	}

	dbContext struct {
		parent context.Context
		db     *sql.DB
		tx     *sql.Tx
	}
)

// New returns a new context with a given DB
func New(parent context.Context, db *sql.DB) Context {
	return &dbContext{
		parent: parent,
		db:     db,
	}
}

// WithTx returns a new context with a given TX
func WithTx(parent context.Context, tx *sql.Tx) Context {
	return &dbContext{
		parent: parent,
		tx:     tx,
	}
}

func (c *dbContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

func (c *dbContext) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *dbContext) Err() error {
	return c.parent.Err()
}

func (c *dbContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

func (c *dbContext) Executor() Executor {
	if c.tx == nil {
		return c.db
	}

	return c.tx
}

func (c *dbContext) Begin(handler Operation) error {
	return c.BeginWith(handler, nil)
}

func (c *dbContext) BeginWith(handler Operation, opts *sql.TxOptions) error {
	var err error
	tx := c.tx
	complete := c.tx == nil

	if tx == nil {
		tx, err = c.db.BeginTx(c, opts)

		if err != nil {
			return err
		}
	}

	err = handler(tx)

	if err != nil {
		if complete {
			tx.Rollback()
		}

		return err
	}

	if complete {
		return tx.Commit()
	}

	return nil
}
