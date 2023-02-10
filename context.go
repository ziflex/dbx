package dbx

import (
	"context"
	"time"
)

type (
	ctxKey struct{}

	defaultContext struct {
		parent   context.Context
		executor Executor
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

// NewContext returns a new context with a given Executor.
func NewContext(parent context.Context, exec Executor) Context {
	return &defaultContext{
		parent:   parent,
		executor: exec,
	}
}

// WithContext returns a new context with a given DB context.
func WithContext(ctx context.Context, dbCtx Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, dbCtx)
}

// FromContext returns a DB context from a given context.
func FromContext(ctx context.Context) Context {
	if dbCtx, ok := ctx.(Context); ok {
		return dbCtx
	}

	if dbCtx, ok := ctx.Value(ctxKey{}).(Context); ok {
		return dbCtx
	}

	return nil
}

// FromContextOrNew returns a DB context from a given context or creates a new one.
func FromContextOrNew(ctx context.Context, db ContextCreator) Context {
	found := FromContext(ctx)

	if found != nil {
		return found
	}

	return db.Context(ctx)
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
	return c.executor
}
