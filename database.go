package dbx

import (
	"context"
	"database/sql"
)

// defaultDatabase implements the Database interface by wrapping a sql.DB instance.
// It provides context-aware database operations and transaction management
// while maintaining full compatibility with the standard database/sql package.
type defaultDatabase struct {
	db *sql.DB
}

// New creates a new Database instance that wraps the provided sql.DB.
// The returned Database provides context-driven database operations and
// automatic transaction management while preserving all the functionality
// of the underlying sql.DB.
//
// The returned database also implements ContextCreator, allowing direct
// context creation via the Context method.
//
// Parameters:
//   - db: A properly initialized sql.DB instance. The caller retains ownership
//     and responsibility for the sql.DB's configuration and driver setup.
//
// Returns:
//   - DatabaseWithContext: A dbx Database that can create contexts and manage transactions.
//
// Example:
//
//	sqlDB, err := sql.Open("postgres", connectionString)
//	if err != nil {
//	    return err
//	}
//	defer sqlDB.Close()
//
//	dbxDB := dbx.New(sqlDB)
//	defer dbxDB.Close()
//
//	ctx := dbxDB.Context(context.Background())
//	rows, err := ctx.Executor().Query("SELECT * FROM users")
func New(db *sql.DB) DatabaseWithContext {
	return &defaultDatabase{db}
}

// Close closes the underlying database connection.
// It's important to call this method when the Database is no longer needed
// to properly release database resources.
func (d *defaultDatabase) Close() error {
	return d.db.Close()
}

// Context creates a new dbx Context from the provided Go context.
// The returned Context embeds this database instance as the executor,
// allowing database operations to be performed within the context's lifecycle.
//
// Parameters:
//   - ctx: The parent Go context for deadline, cancellation, and value propagation.
//
// Returns:
//   - Context: A dbx Context that can be used for database operations.
func (d *defaultDatabase) Context(ctx context.Context) Context {
	return NewContext(ctx, d)
}

// Begin starts a transaction with default options.
// This method delegates to the underlying sql.DB's Begin method.
func (d *defaultDatabase) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// BeginTx starts a transaction with the provided context and options.
// The context is used for the transaction's lifecycle - if the context
// is canceled, the transaction will be rolled back.
//
// Parameters:
//   - ctx: Context for the transaction's lifecycle
//   - opts: Transaction options including isolation level and read-only flag
func (d *defaultDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

// Exec executes a query without returning any rows.
// This method delegates to the underlying sql.DB's Exec method.
//
// Parameters:
//   - query: SQL query string, potentially with placeholder parameters
//   - args: Arguments for placeholder parameters in the query
//
// Returns:
//   - sql.Result: Contains information about the query execution (rows affected, last insert ID)
//   - error: Any error that occurred during query execution
func (d *defaultDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Query executes a query that returns rows, typically a SELECT statement.
// This method delegates to the underlying sql.DB's Query method.
//
// Parameters:
//   - query: SQL query string, potentially with placeholder parameters
//   - args: Arguments for placeholder parameters in the query
//
// Returns:
//   - *sql.Rows: Rows returned by the query. Must be closed after use.
//   - error: Any error that occurred during query execution
func (d *defaultDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// This method delegates to the underlying sql.DB's QueryRow method.
// QueryRow always returns a non-nil value; errors are deferred until Row's Scan method is called.
//
// Parameters:
//   - query: SQL query string, potentially with placeholder parameters
//   - args: Arguments for placeholder parameters in the query
//
// Returns:
//   - *sql.Row: Single row result. Use Scan() to extract values and check for errors.
func (d *defaultDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// ExecContext executes a query without returning any rows, using the provided context.
// This method delegates to the underlying sql.DB's ExecContext method.
// The context can be used to cancel the query execution if it takes too long.
//
// Parameters:
//   - dbContext: Context for controlling query execution lifecycle
//   - query: SQL query string, potentially with placeholder parameters
//   - args: Arguments for placeholder parameters in the query
//
// Returns:
//   - sql.Result: Contains information about the query execution (rows affected, last insert ID)
//   - error: Any error that occurred during query execution or context cancellation
func (d *defaultDatabase) ExecContext(dbContext context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(dbContext, query, args...)
}

// QueryContext executes a query that returns rows, using the provided context.
// This method delegates to the underlying sql.DB's QueryContext method.
// The context can be used to cancel the query execution if it takes too long.
//
// Parameters:
//   - dbContext: Context for controlling query execution lifecycle
//   - query: SQL query string, potentially with placeholder parameters
//   - args: Arguments for placeholder parameters in the query
//
// Returns:
//   - *sql.Rows: Rows returned by the query. Must be closed after use.
//   - error: Any error that occurred during query execution or context cancellation
func (d *defaultDatabase) QueryContext(dbContext context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(dbContext, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row, using the provided context.
// This method delegates to the underlying sql.DB's QueryRowContext method.
// QueryRowContext always returns a non-nil value; errors are deferred until Row's Scan method is called.
//
// Parameters:
//   - dbContext: Context for controlling query execution lifecycle
//   - query: SQL query string, potentially with placeholder parameters
//   - args: Arguments for placeholder parameters in the query
//
// Returns:
//   - *sql.Row: Single row result. Use Scan() to extract values and check for errors.
func (d *defaultDatabase) QueryRowContext(dbContext context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(dbContext, query, args...)
}
