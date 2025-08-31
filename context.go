package dbx

import (
	"context"
	"time"
)

type (
	// ctxKey is an unexported type used as a key for storing dbx Context
	// instances in Go contexts. Using an unexported type prevents collisions
	// with other packages that might use context values.
	ctxKey struct{}

	// defaultContext implements the Context interface by wrapping a parent
	// Go context and embedding a database executor. It delegates all standard
	// context operations to the parent while providing access to the database executor.
	defaultContext struct {
		parent   context.Context // Parent Go context for standard context operations
		executor Executor        // Database executor (sql.DB or sql.Tx)
	}
)

// Is returns true if the provided context is a dbx Context.
// This function performs a type assertion to check if the context
// implements the dbx Context interface.
//
// Parameters:
//   - ctx: The context to check
//
// Returns:
//   - bool: true if ctx is a dbx Context, false otherwise
//
// Example:
//
//	if dbx.Is(ctx) {
//	    // Safe to use dbx-specific operations
//	    dbCtx, _ := dbx.As(ctx)
//	    executor := dbCtx.Executor()
//	}
func Is(ctx context.Context) bool {
	_, ok := ctx.(Context)

	return ok
}

// As attempts to extract a dbx Context from the provided context.
// This function performs a type assertion and returns both the Context
// and a boolean indicating success.
//
// Parameters:
//   - ctx: The context to extract dbx Context from
//
// Returns:
//   - Context: The extracted dbx Context (nil if extraction failed)
//   - bool: true if extraction was successful, false otherwise
//
// Example:
//
//	if dbCtx, ok := dbx.As(ctx); ok {
//	    executor := dbCtx.Executor()
//	    // Use executor for database operations
//	}
func As(ctx context.Context) (Context, bool) {
	c, ok := ctx.(Context)

	return c, ok
}

// NewContext creates a new dbx Context by combining a parent Go context
// with a database executor. The resulting Context can be used for database
// operations while preserving the parent context's deadline, cancellation,
// and value propagation behavior.
//
// Parameters:
//   - parent: The parent Go context to wrap
//   - exec: The database executor (sql.DB or sql.Tx) to embed
//
// Returns:
//   - Context: A new dbx Context combining the parent context and executor
//
// Example:
//
//	sqlDB, _ := sql.Open("postgres", connectionString)
//	dbCtx := dbx.NewContext(context.Background(), sqlDB)
//	result, err := dbCtx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
func NewContext(parent context.Context, exec Executor) Context {
	return &defaultContext{
		parent:   parent,
		executor: exec,
	}
}

// NewDatabaseContext creates a new dbx Context from a Database instance.
// This is a convenience function that allows creating a context from any Database,
// regardless of whether it implements ContextCreator or not.
//
// Parameters:
//   - parent: The parent Go context to wrap
//   - db: The Database instance to use as the executor
//
// Returns:
//   - Context: A new dbx Context with the database as executor
//
// Example:
//
//	db := dbx.New(sqlDB)
//	dbCtx := dbx.NewDatabaseContext(context.Background(), db)
//	result, err := dbCtx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
func NewDatabaseContext(parent context.Context, db Database) Context {
	return &defaultContext{
		parent:   parent,
		executor: db,
	}
}

// NewContextFrom attempts to find an existing dbx Context in the provided context,
// or creates a new one using the provided creator if none is found.
// This function is useful for ensuring that a dbx Context is available while
// avoiding unnecessary Context creation when one already exists.
//
// The creator parameter can be either a ContextCreator, a Database, or any type that has
// a Context(context.Context) Context method.
//
// Parameters:
//   - ctx: The context to search for an existing dbx Context
//   - creator: Either a ContextCreator, Database, or Transactor
//
// Returns:
//   - Context: Either the existing dbx Context or a newly created one
//
// Example:
//
//	// This will reuse existing dbx Context or create new one
//	dbCtx := dbx.NewContextFrom(ctx, database)
//	executor := dbCtx.Executor()
func NewContextFrom(ctx context.Context, input any) Context {
	found := FromContext(ctx)

	if found != nil {
		return found
	}

	switch val := input.(type) {
	case ContextCreator:
		return val.Context(ctx)
	case Database:
		return NewDatabaseContext(ctx, val)
	case Transactor:
		return NewContext(ctx, val)
	default:
		// If none work, panic with helpful message
		panic("input must implement ContextCreator, Database, or Transactor")
	}
}

// FromContext extracts a dbx Context from the provided Go context.
// It first checks if the context itself is a dbx Context, then checks
// if a dbx Context is stored as a value within the context using the
// internal context key.
//
// Parameters:
//   - ctx: The context to extract dbx Context from
//
// Returns:
//   - Context: The extracted dbx Context, or nil if none found
//
// Example:
//
//	dbCtx := dbx.FromContext(ctx)
//	if dbCtx != nil {
//	    executor := dbCtx.Executor()
//	    // Use executor for database operations
//	}
func FromContext(ctx context.Context) Context {
	if dbCtx, ok := ctx.(Context); ok {
		return dbCtx
	}

	if dbCtx, ok := ctx.Value(ctxKey{}).(Context); ok {
		return dbCtx
	}

	return nil
}

// WithContext embeds a dbx Context into a Go context as a value.
// This allows the dbx Context to be retrieved later using FromContext.
// This is useful when you need to pass a dbx Context through code that
// expects a standard Go context.
//
// Parameters:
//   - ctx: The parent Go context
//   - dbCtx: The dbx Context to embed
//
// Returns:
//   - context.Context: A new context containing the embedded dbx Context
//
// Example:
//
//	dbCtx := database.Context(ctx)
//	embeddedCtx := dbx.WithContext(context.Background(), dbCtx)
//	// Later: extractedCtx := dbx.FromContext(embeddedCtx)
func WithContext(ctx context.Context, dbCtx Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, dbCtx)
}

// Deadline delegates to the parent context's Deadline method.
// Returns the time when work done on behalf of this context should be canceled.
func (c *defaultContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

// Done delegates to the parent context's Done method.
// Returns a channel that's closed when work done on behalf of this context should be canceled.
func (c *defaultContext) Done() <-chan struct{} {
	return c.parent.Done()
}

// Err delegates to the parent context's Err method.
// Returns a non-nil error value after Done is closed, indicating why the context was canceled.
func (c *defaultContext) Err() error {
	return c.parent.Err()
}

// Value delegates to the parent context's Value method.
// Returns the value associated with this context for key, or nil if no value is associated with key.
func (c *defaultContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

// Executor returns the database executor embedded in this context.
// The executor can be either a sql.DB (for direct database operations)
// or a sql.Tx (when operating within a transaction).
func (c *defaultContext) Executor() Executor {
	return c.executor
}
