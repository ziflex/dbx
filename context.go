package dbx

import (
	"context"
	"database/sql"
	"time"
)

type DefaultContext struct {
	parent context.Context
	db     *sql.DB
	tx     *sql.Tx
}

// With returns a new context with a given DB context.
func With(ctx context.Context, dbCtx Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, dbCtx)
}

// From returns a DB context from a given context.
func From(ctx context.Context) Context {
	if dbCtx, ok := ctx.Value(ctxKey{}).(Context); ok {
		return dbCtx
	}

	return nil
}

// FromOrNew returns a DB context from a given context or creates a new one.
func FromOrNew(ctx context.Context, db Database) Context {
	if dbCtx, ok := ctx.Value(ctxKey{}).(Context); ok {
		return dbCtx
	}

	return db.Context(ctx)
}

// FromDB returns a new context with a given DB
func FromDB(parent context.Context, db *sql.DB) Context {
	return &DefaultContext{
		parent: parent,
		db:     db,
	}
}

// FromTx returns a new context with a given TX
func FromTx(parent context.Context, tx *sql.Tx) Context {
	return &DefaultContext{
		parent: parent,
		tx:     tx,
	}
}

func (c *DefaultContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

func (c *DefaultContext) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *DefaultContext) Err() error {
	return c.parent.Err()
}

func (c *DefaultContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

func (c *DefaultContext) Executor() Executor {
	if c.tx == nil {
		return c.db
	}

	return c.tx
}
