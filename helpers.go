package dbx

import "context"

type ctxKey struct{}

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
