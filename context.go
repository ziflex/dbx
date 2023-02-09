package dbx

import (
	"context"
	"database/sql"
	"time"
)

type (
	ctxKey struct{}

	defaultContext struct {
		parent context.Context
		db     *sql.DB
		tx     *sql.Tx
	}
)

// Is returns true if the context is a DB context.
func Is(ctx context.Context) bool {
	_, ok := ctx.(Context)

	return ok
}

// As returns a DB context if the context is a DB context.
func As(ctx context.Context) (Context, bool) {
	c, ok := ctx.(Context)

	return c, ok
}

// With returns a new context with a given DB context.
func With(ctx context.Context, dbCtx Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, dbCtx)
}

// From returns a DB context from a given context.
func From(ctx context.Context) Context {
	if dbCtx, ok := ctx.(Context); ok {
		return dbCtx
	}

	if dbCtx, ok := ctx.Value(ctxKey{}).(Context); ok {
		return dbCtx
	}

	return nil
}

// FromOrNew returns a DB context from a given context or creates a new one.
func FromOrNew(ctx context.Context, db Database) Context {
	found := From(ctx)

	if found != nil {
		return found
	}

	return db.Context(ctx)
}

// FromDB returns a new context with a given DB
func FromDB(parent context.Context, db *sql.DB) Context {
	return &defaultContext{
		parent: parent,
		db:     db,
	}
}

// FromTx returns a new context with a given TX
func FromTx(parent context.Context, tx *sql.Tx) Context {
	return &defaultContext{
		parent: parent,
		tx:     tx,
	}
}

func (c *defaultContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

func (c *defaultContext) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *defaultContext) Err() error {
	return c.parent.Err()
}

func (c *defaultContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

func (c *defaultContext) Executor() Executor {
	if c.tx == nil {
		return c.db
	}

	return c.tx
}
