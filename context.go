package dbx

import (
	"context"
	"time"
)

type (
	ctxKey struct{}

	DefaultContext struct {
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
	return &DefaultContext{
		parent:   parent,
		executor: exec,
	}
}

// NewContextFrom returns a DB context from a given context or creates a new one if an existing one not found in a given context.
func NewContextFrom(ctx context.Context, creator ContextCreator) Context {
	found := FromContext(ctx)

	if found != nil {
		return found
	}

	return creator.Context(ctx)
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

// WithContext returns a new context with a given DB context.
func WithContext(ctx context.Context, dbCtx Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, dbCtx)
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
	return c.executor
}
